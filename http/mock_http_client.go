package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	StatusOK = http.StatusOK
)

type mockBody struct {
	body   string
	reader io.Reader
}

func (m *mockBody) Read(p []byte) (n int, err error) {
	return m.reader.Read(p)
}

func (m *mockBody) Close() error {
	return nil
}

type MockHttpClient struct {
	bodyMocks  map[string]string
	errorMocks map[string]error
}

func (m *MockHttpClient) key(method, url string) string {
	return method + "|" + url
}

func (m *MockHttpClient) WillReturnBody(method, url string, body string) {
	if m.bodyMocks == nil {
		m.bodyMocks = make(map[string]string)
	}
	m.bodyMocks[m.key(method, url)] = body
}

func (m *MockHttpClient) WillReturnError(method, url string, err error) {
	if m.errorMocks == nil {
		m.errorMocks = make(map[string]error)
	}
	m.errorMocks[m.key(method, url)] = err
}

func (m *MockHttpClient) Get(url string) (resp *http.Response, err error) {
	return m.doRequest("GET", url)
}

func (m *MockHttpClient) Post(url string) (resp *http.Response, err error) {
	return m.doRequest("POST", url)
}

func (m *MockHttpClient) doRequest(method, url string) (resp *http.Response, err error) {
	err = m.errorMocks[m.key(method, url)]

	if err != nil {
		return nil, err
	}

	body, ok := m.bodyMocks[m.key(method, url)]
	if !ok {
		return nil, fmt.Errorf("no body mock found for %s %s", method, url)
	}

	return &http.Response{
		Body:       &mockBody{body: body, reader: strings.NewReader(body)},
		StatusCode: StatusOK,
	}, nil
}

func (m *MockHttpClient) NewRequest(method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	key := m.key(req.Method, req.URL.String())

	body, exists := m.bodyMocks[key]
	if !exists {
		return nil, http.ErrNoLocation
	}

	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}

	return resp, nil
}
