package httpro_test

import (
	"net/http"
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

var _ = Describe("httpro", func() {
	var (
		err error
		res *http.Response
		ts  *testServer
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
		Describe("standard route", func() {
			Describe("default client", func() {
				BeforeEach(func() {
					ts = newTestServer(testServerConfig{})
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

		Describe("connection refused", func() {
			Describe("default client", func() {
				BeforeEach(func() {
					ts = newTestServer(testServerConfig{NoConnectRefusedAfter: 300 * time.Millisecond})
					c := httpro.NewHTTPClient(httpro.Config{})
					res, err = c.Get(ts.URL)
				})

				It("should respond with connect refused error", func() {
					Expect(httpro.IsErrConnectionRefused(err)).To(BeTrue())
				})

				It("should respond empty", func() {
					Expect(res).To(BeNil())
				})
			})

			Describe("reconnect delay client", func() {
				BeforeEach(func() {
					ts = newTestServer(testServerConfig{NoConnectRefusedAfter: 300 * time.Millisecond})
					c := httpro.NewHTTPClient(httpro.Config{ReconnectDelay: 400 * time.Millisecond})
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

		Describe("request timed out", func() {
			Describe("default client", func() {
				BeforeEach(func() {
					ts = newTestServer(testServerConfig{NoRequestTimeoutOnRetry: 3})
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

			Describe("request timeout route", func() {
				BeforeEach(func() {
					ts = newTestServer(testServerConfig{NoRequestTimeoutOnRetry: 3})
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

		Describe("request timed out and request retry", func() {
			Describe("request timed out route", func() {
				BeforeEach(func() {
					ts = newTestServer(testServerConfig{NoRequestTimeoutOnRetry: 3})
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

			Describe("request timed out route; enough retries", func() {
				BeforeEach(func() {
					ts = newTestServer(testServerConfig{NoRequestTimeoutOnRetry: 3})
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
