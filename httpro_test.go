package httpro_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/giantswarm/httpro"
)

func TestHTTPro(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "httpro")
}

type testServerConfig struct {
	NoRequestTimeoutOnRetry uint
	NoConnectTimeoutAfter   time.Duration
}

func testServerWithRequestTimeout(c testServerConfig) *httptest.Server {
	reqCount := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCount++

		if reqCount < int(c.NoRequestTimeoutOnRetry) {
			time.Sleep(100 * time.Millisecond)
		}

		fmt.Fprint(w, "OK")
	}))

	return ts
}

var _ = Describe("httpro", func() {
	var (
		err error
		res *http.Response
		ts  *httptest.Server
	)

	BeforeEach(func() {
		err = nil
		ts = nil
	})

	AfterEach(func() {
		if ts != nil {
			ts.Close()
		}

		if res != nil {
			res.Body.Close()
		}
	})

	Describe("GET", func() {
		Describe("default client", func() {
			Describe("standard route", func() {
				BeforeEach(func() {
					ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						fmt.Fprint(w, "OK")
					}))

					c := httpro.NewHTTPClient(httpro.Config{})
					res, err = c.Get(ts.URL)
				})

				It("should respond without error", func() {
					Expect(err).To(BeNil())
				})

				It("should respond with status code 200", func() {
					Expect(res.StatusCode).To(Equal(200))
				})
			})

			Describe("request-timeout route", func() {
				BeforeEach(func() {
					ts = testServerWithRequestTimeout(testServerConfig{NoRequestTimeoutOnRetry: 3})
					c := httpro.NewHTTPClient(httpro.Config{})
					res, err = c.Get(ts.URL)
				})

				It("should respond without error", func() {
					Expect(err).To(BeNil())
				})

				It("should respond with status code 200", func() {
					Expect(res.StatusCode).To(Equal(200))
				})
			})
		})

		Describe("request-timeout client", func() {
			Describe("request-timeout route", func() {
				BeforeEach(func() {
					ts = testServerWithRequestTimeout(testServerConfig{NoRequestTimeoutOnRetry: 3})
					c := httpro.NewHTTPClient(httpro.Config{RequestTimeout: 50 * time.Millisecond})
					res, err = c.Get(ts.URL)
				})

				It("should respond with timeout error", func() {
					Expect(httpro.IsErrRequestTimeout(err)).To(BeTrue())
				})

				It("should respond empty", func() {
					Expect(res).To(BeNil())
				})
			})
		})

		Describe("request-timeout and request-retry client", func() {
			Describe("request-timeout route", func() {
				BeforeEach(func() {
					ts = testServerWithRequestTimeout(testServerConfig{NoRequestTimeoutOnRetry: 3})
					c := httpro.NewHTTPClient(httpro.Config{RequestTimeout: 50 * time.Millisecond, RequestRetry: 2})
					res, err = c.Get(ts.URL)
				})

				It("should respond with timeout error", func() {
					Expect(httpro.IsErrRequestTimeout(err)).To(BeTrue())
				})

				It("should respond empty", func() {
					Expect(res).To(BeNil())
				})
			})

			Describe("request-timeout route; enough retries", func() {
				BeforeEach(func() {
					ts = testServerWithRequestTimeout(testServerConfig{NoRequestTimeoutOnRetry: 3})
					c := httpro.NewHTTPClient(httpro.Config{RequestTimeout: 50 * time.Millisecond, RequestRetry: 3})
					res, err = c.Get(ts.URL)
				})

				It("should respond without error", func() {
					Expect(err).To(BeNil())
				})

				It("should respond with status code 200", func() {
					Expect(res.StatusCode).To(Equal(200))
				})
			})
		})
	})
})
