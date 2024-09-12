package http

import (
	"io"
	"net/http"
)

type Client interface {
	Get(url string) (resp *http.Response, err error)
	NewRequest(method string, url string, body io.Reader) (*http.Request, error)
	Do(req *http.Request) (*http.Response, error)
}
