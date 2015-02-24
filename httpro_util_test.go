package httpro_test

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

type testServerConfig struct {
	Handler               http.HandlerFunc
	NoConnectRefusedAfter time.Duration
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

func newTestServerAddr() string {
	var addr string

	// Check if the given port is available. It it is not, try again.
	tl, terr := net.Listen("tcp", "127.0.0.1:0")
	if terr != nil {
		return newTestServerAddr()
	} else {
		addr = tl.Addr().String()
		tl.Close()
	}

	return addr
}

func newTestHTTPHandlerRequestTimeoutRetry(timeout time.Duration, noTimeoutOnRetry int) http.HandlerFunc {
	var retry int64

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retry = atomic.AddInt64(&retry, 1)

		if int(retry) < noTimeoutOnRetry {
			time.Sleep(timeout)
		}

		fmt.Fprint(w, "OK")
	})
}

func newTestHTTPHandlerStatusCodeRetry(statusCode, noCodeOnRetry int) http.HandlerFunc {
	retry := 0

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retry++

		if retry < noCodeOnRetry {
			w.WriteHeader(statusCode)
			fmt.Fprint(w, strconv.Itoa(statusCode))
		}

		fmt.Fprint(w, "OK")
	})
}

func newTestServer(c testServerConfig) *testServer {
	addr := newTestServerAddr()

	var l net.Listener
	ts := &testServer{
		URL:      "http://" + addr,
		Listener: l,
		Server: &http.Server{
			Handler: c.Handler,
		},
	}

	var err error
	ts.Listener, err = net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	go func() {
		ts.Listen()
	}()

	return ts
}

func newTestServerConnectDelay(c testServerConfig) *testServer {
	addr := newTestServerAddr()

	var l net.Listener
	ts := &testServer{
		URL:      "http://" + addr,
		Listener: l,
		Server: &http.Server{
			Handler: c.Handler,
		},
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
