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
	n := random(45000, 49000)
	port := strconv.Itoa(n)

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
		URL:      "http://localhost:" + port,
		Listener: l,
		Server:   srv,
	}

	go func() {
		time.Sleep(c.NoConnectRefusedAfter)

		var err error
		ts.Listener, err = net.Listen("tcp", "localhost:"+port)
		if err != nil {
			panic(err)
		}

		ts.Listen()
	}()

	return ts
}
