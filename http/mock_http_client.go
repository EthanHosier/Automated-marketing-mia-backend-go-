package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
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
	bodyMocks   []bodyMock
	errorMocks  []errorMock
	headerMocks []headerMock
}

type bodyMock struct {
	method string
	url    *regexp.Regexp
	body   string
}

type errorMock struct {
	method string
	url    *regexp.Regexp
	err    error
}

type headerMock struct {
	method  string
	url     *regexp.Regexp
	headers map[string]string
}

func (m *MockHttpClient) WillReturnBody(method, urlPattern, body string) {
	if m.bodyMocks == nil {
		m.bodyMocks = make([]bodyMock, 0)
	}

	// Compile the regex pattern
	re, err := regexp.Compile(urlPattern)
	if err != nil {
		panic(err) // Handle error appropriately
	}

	// Check if we already have a mock for this method and URL
	for i, bm := range m.bodyMocks {
		if bm.method == method && bm.url.MatchString(urlPattern) {
			m.bodyMocks[i] = bodyMock{method: method, url: re, body: body}
			return
		}
	}

	// Add a new mock if not found
	m.bodyMocks = append(m.bodyMocks, bodyMock{method: method, url: re, body: body})
}

func (m *MockHttpClient) WillReturnError(method, urlPattern string, err error) {
	if m.errorMocks == nil {
		m.errorMocks = make([]errorMock, 0)
	}

	// Compile the regex pattern
	re, err := regexp.Compile(urlPattern)
	if err != nil {
		panic(err) // Handle error appropriately
	}

	// Check if we already have a mock for this method and URL
	for i, em := range m.errorMocks {
		if em.method == method && em.url.MatchString(urlPattern) {
			m.errorMocks[i] = errorMock{method: method, url: re, err: err}
			return
		}
	}

	// Add a new mock if not found
	m.errorMocks = append(m.errorMocks, errorMock{method: method, url: re, err: err})
}

func (m *MockHttpClient) WillReturnHeader(method, urlPattern string, headers map[string]string) {
	if m.headerMocks == nil {
		m.headerMocks = make([]headerMock, 0)
	}

	// Compile the regex pattern
	re, err := regexp.Compile(urlPattern)
	if err != nil {
		panic(err) // Handle error appropriately
	}

	// Check if we already have a mock for this method and URL
	for i, hm := range m.headerMocks {
		if hm.method == method && hm.url.MatchString(urlPattern) {
			m.headerMocks[i] = headerMock{method: method, url: re, headers: headers}
			return
		}
	}

	// Add a new mock if not found
	m.headerMocks = append(m.headerMocks, headerMock{method: method, url: re, headers: headers})
}

func (m *MockHttpClient) Get(url string) (resp *http.Response, err error) {
	return m.doRequest("GET", url)
}

func (m *MockHttpClient) Post(url string) (resp *http.Response, err error) {
	return m.doRequest("POST", url)
}

func (m *MockHttpClient) doRequest(method, url string) (resp *http.Response, err error) {
	for _, em := range m.errorMocks {
		if em.method == method && em.url.MatchString(url) {
			return nil, em.err
		}
	}

	var body string
	match := false
	for _, bm := range m.bodyMocks {
		if bm.method == method && bm.url.MatchString(url) {
			body = bm.body
			match = true
			break
		}
	}
	if !match {
		return nil, fmt.Errorf("no body mock found for %s %s", method, url)
	}

	headers := make(http.Header)
	for _, hm := range m.headerMocks {
		if hm.method == method && hm.url.MatchString(url) {
			for k, v := range hm.headers {
				headers.Set(k, v)
			}
			break
		}
	}

	return &http.Response{
		Body:       &mockBody{body: body, reader: strings.NewReader(body)},
		StatusCode: StatusOK,
		Header:     headers,
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
	// Use the regex pattern matching for the request method and URL
	for _, bm := range m.bodyMocks {
		if bm.method == req.Method && bm.url.MatchString(req.URL.String()) {
			body := bm.body
			var headers = make(http.Header)
			for _, hm := range m.headerMocks {
				if hm.method == req.Method && hm.url.MatchString(req.URL.String()) {
					for k, v := range hm.headers {
						headers.Set(k, v)
					}
					break
				}
			}
			resp := &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
				Header:     headers,
			}
			return resp, nil
		}
	}

	return nil, errors.New("no body mock found")
}
