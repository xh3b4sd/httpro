package httpro_test

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"
)

type testServerConfig struct {
	StatusCode              int
	NoRequestTimeoutOnRetry uint
	NoStatusCodeOnRetry     uint
	NoConnectRefusedAfter   time.Duration
}

type testServer struct {
	URL      string
	Listener net.Listener
	Server   *http.Server
}

func (ts *testServer) Close() {
	ts.Listener.Close()
}

func (ts *testServer) Listen() {
	ts.Server.Serve(ts.Listener)
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func newTestServer(c testServerConfig) *testServer {
	var addr string
	// Check if the given port is available. It it is not, try again.
	tl, terr := net.Listen("tcp", "127.0.0.1:0")
	if terr != nil {
		return newTestServer(c)
	} else {
		addr = tl.Addr().String()
		tl.Close()
	}

	reqCount := 0

	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqCount++

			if c.StatusCode > 0 && c.NoStatusCodeOnRetry > 0 && reqCount < int(c.NoStatusCodeOnRetry) {
				w.WriteHeader(c.StatusCode)
				fmt.Fprint(w, strconv.Itoa(c.StatusCode))
				return
			}

			if c.NoRequestTimeoutOnRetry > 0 && reqCount < int(c.NoRequestTimeoutOnRetry) {
				time.Sleep(100 * time.Millisecond)
			}

			fmt.Fprint(w, "OK")
		}),
	}

	var l net.Listener
	ts := &testServer{
		URL:      "http://" + addr,
		Listener: l,
		Server:   srv,
	}

	go func() {
		time.Sleep(c.NoConnectRefusedAfter)

		var err error
		ts.Listener, err = net.Listen("tcp", addr)
		if err != nil {
			panic(err)
		}

		ts.Listen()
	}()

	return ts
}
