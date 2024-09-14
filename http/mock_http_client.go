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

func (m *MockHttpClient) Get(url string) (resp *http.Response, err error) {
	return m.doRequest("GET", url)
}

func (m *MockHttpClient) Post(url string) (resp *http.Response, err error) {
	return m.doRequest("POST", url)
}

// Exact URL mock methods
func (m *MockHttpClient) WillReturnBody(method, url string, body string) {
	if m.bodyMocks == nil {
		m.bodyMocks = make([]bodyMock, 0)
	}

	for i, bm := range m.bodyMocks {
		if bm.method == method && bm.url.MatchString(url) {
			m.bodyMocks[i] = bodyMock{method: method, url: regexp.MustCompile("^" + regexp.QuoteMeta(url) + "$"), body: body}
			return
		}
	}

	m.bodyMocks = append(m.bodyMocks, bodyMock{method: method, url: regexp.MustCompile("^" + regexp.QuoteMeta(url) + "$"), body: body})
}

func (m *MockHttpClient) WillReturnError(method, url string, err error) {
	if m.errorMocks == nil {
		m.errorMocks = make([]errorMock, 0)
	}

	for i, em := range m.errorMocks {
		if em.method == method && em.url.MatchString(url) {
			m.errorMocks[i] = errorMock{method: method, url: regexp.MustCompile("^" + regexp.QuoteMeta(url) + "$"), err: err}
			return
		}
	}

	m.errorMocks = append(m.errorMocks, errorMock{method: method, url: regexp.MustCompile("^" + regexp.QuoteMeta(url) + "$"), err: err})
}

func (m *MockHttpClient) WillReturnHeader(method, url string, headers map[string]string) {
	if m.headerMocks == nil {
		m.headerMocks = make([]headerMock, 0)
	}

	for i, hm := range m.headerMocks {
		if hm.method == method && hm.url.MatchString(url) {
			m.headerMocks[i] = headerMock{method: method, url: regexp.MustCompile("^" + regexp.QuoteMeta(url) + "$"), headers: headers}
			return
		}
	}

	m.headerMocks = append(m.headerMocks, headerMock{method: method, url: regexp.MustCompile("^" + regexp.QuoteMeta(url) + "$"), headers: headers})
}

// Regex-based mock methods
func (m *MockHttpClient) WillReturnBodyRegex(method, urlPattern, body string) {
	if m.bodyMocks == nil {
		m.bodyMocks = make([]bodyMock, 0)
	}

	re, err := regexp.Compile(urlPattern)
	if err != nil {
		panic(err) // Handle error appropriately
	}

	for i, bm := range m.bodyMocks {
		if bm.method == method && bm.url.String() == re.String() {
			m.bodyMocks[i] = bodyMock{method: method, url: re, body: body}
			return
		}
	}

	m.bodyMocks = append(m.bodyMocks, bodyMock{method: method, url: re, body: body})
}

func (m *MockHttpClient) WillReturnErrorRegex(method, urlPattern string, err error) {
	if m.errorMocks == nil {
		m.errorMocks = make([]errorMock, 0)
	}

	re, err := regexp.Compile(urlPattern)
	if err != nil {
		panic(err) // Handle error appropriately
	}

	for i, em := range m.errorMocks {
		if em.method == method && em.url.String() == re.String() {
			m.errorMocks[i] = errorMock{method: method, url: re, err: err}
			return
		}
	}

	m.errorMocks = append(m.errorMocks, errorMock{method: method, url: re, err: err})
}

func (m *MockHttpClient) WillReturnHeaderRegex(method, urlPattern string, headers map[string]string) {
	if m.headerMocks == nil {
		m.headerMocks = make([]headerMock, 0)
	}

	re, err := regexp.Compile(urlPattern)
	if err != nil {
		panic(err) // Handle error appropriately
	}

	for i, hm := range m.headerMocks {
		if hm.method == method && hm.url.String() == re.String() {
			m.headerMocks[i] = headerMock{method: method, url: re, headers: headers}
			return
		}
	}

	m.headerMocks = append(m.headerMocks, headerMock{method: method, url: re, headers: headers})
}

// HTTP request handling
func (m *MockHttpClient) doRequest(method, url string) (resp *http.Response, err error) {
	// Check for errors first with exact matches
	for _, em := range m.errorMocks {
		if em.method == method && em.url.MatchString(url) {
			return nil, em.err
		}
	}

	// Check for exact body match
	for _, bm := range m.bodyMocks {
		if bm.method == method && bm.url.MatchString(url) {
			body := bm.body
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
	}

	// Check for regex matches if no exact match was found
	for _, bm := range m.bodyMocks {
		if bm.method == method && bm.url.MatchString(url) {
			body := bm.body
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
	}

	return nil, fmt.Errorf("no body mock found for %s %s", method, url)
}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	// First, check for exact URL matches
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
			return &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
				Header:     headers,
			}, nil
		}
	}

	// Then check for regex matches if no exact match was found
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
			return &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
				Header:     headers,
			}, nil
		}
	}

	return nil, errors.New("no body mock found")
}

func (m *MockHttpClient) NewRequest(method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	return req, nil
}
