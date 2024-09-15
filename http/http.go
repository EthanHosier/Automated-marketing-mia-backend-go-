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

type HttpClient struct {
}

func (c *HttpClient) Get(url string) (resp *http.Response, err error) {
	return http.Get(url)
}

func (c *HttpClient) NewRequest(method string, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}

func (c *HttpClient) Do(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}
