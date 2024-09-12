package http

import (
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

func (m *MockHttpClient) WillReturnBody(url string, body string) {
	if m.bodyMocks == nil {
		m.bodyMocks = make(map[string]string)
	}
	m.bodyMocks[url] = body
}

func (m *MockHttpClient) WillReturnError(url string, err error) {
	if m.errorMocks == nil {
		m.errorMocks = make(map[string]error)
	}
	m.errorMocks[url] = err
}

func (m *MockHttpClient) Get(url string) (resp *http.Response, err error) {
	err = m.errorMocks[url]

	if err != nil {
		return nil, err
	}

	body, ok := m.bodyMocks[url]
	if !ok {
		return nil, fmt.Errorf("no body mock found for %s", url)
	}

	return &http.Response{
		Body:       &mockBody{body: body, reader: strings.NewReader(body)},
		StatusCode: StatusOK,
	}, nil
}
