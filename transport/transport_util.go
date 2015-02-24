package transport

import (
	"net/http"
)

type context struct {
	req *http.Request
	res *http.Response
}

func isStatusCode5XX(statusCode int) bool {
	return statusCode >= 500 && statusCode <= 599
}
