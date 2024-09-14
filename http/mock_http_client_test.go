package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestMockHttpClient_Get_Success(t *testing.T) {
	client := &MockHttpClient{}
	mockURL := "http://example.com"
	mockBody := `{"message": "success"}`
	client.WillReturnBody("GET", mockURL, mockBody)

	resp, err := client.Get(mockURL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != StatusOK {
		t.Errorf("expected status code 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if string(body) != mockBody {
		t.Errorf("expected body %s, got %s", mockBody, string(body))
	}
}

func TestMockHttpClient_Get_NoMock(t *testing.T) {
	client := &MockHttpClient{}
	mockURL := "http://example.com"

	_, err := client.Get(mockURL)

	if err == nil {
		t.Fatalf("expected an error, got none")
	}

	expectedError := fmt.Sprintf("no body mock found for GET %s", mockURL)
	if err.Error() != expectedError {
		t.Errorf("expected error %s, got %s", expectedError, err.Error())
	}
}

func TestMockHttpClient_Get_MultipleURLs(t *testing.T) {
	client := &MockHttpClient{}
	client.WillReturnBody("GET", "http://example.com/1", `{"message": "success1"}`)
	client.WillReturnBody("GET", "http://example.com/2", `{"message": "success2"}`)

	resp1, err := client.Get("http://example.com/1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp1.Body.Close()

	body1, err := io.ReadAll(resp1.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	if string(body1) != `{"message": "success1"}` {
		t.Errorf("expected body %s, got %s", `{"message": "success1"}`, string(body1))
	}

	resp2, err := client.Get("http://example.com/2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp2.Body.Close()

	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	if string(body2) != `{"message": "success2"}` {
		t.Errorf("expected body %s, got %s", `{"message": "success2"}`, string(body2))
	}
}

func TestMockHttpClient_Get_EmptyBody(t *testing.T) {
	client := &MockHttpClient{}
	mockURL := "http://example.com/empty"
	client.WillReturnBody("GET", mockURL, "")

	resp, err := client.Get(mockURL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if string(body) != "" {
		t.Errorf("expected empty body, got %s", string(body))
	}
}

func TestMockHttpClient_Get_OverwriteBody(t *testing.T) {
	client := &MockHttpClient{}
	mockURL := "http://example.com"
	client.WillReturnBody("GET", mockURL, `{"message": "old"}`)
	client.WillReturnBody("GET", mockURL, `{"message": "new"}`)

	resp, err := client.Get(mockURL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	expectedBody := `{"message": "new"}`
	if string(body) != expectedBody {
		t.Errorf("expected body %s, got %s", expectedBody, string(body))
	}
}

func TestMockHttpClient_WillReturnError_Success(t *testing.T) {
	client := &MockHttpClient{}
	mockURL := "http://example.com"
	mockError := fmt.Errorf("mock error")
	client.WillReturnError("POST", mockURL, mockError)
	expectedErr := "no body mock found for GET http://example.com"

	_, err := client.Get(mockURL)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}

	if err.Error() == mockError.Error() {
		t.Errorf("expected error %v, got %v", mockError, err)
	}

	if err.Error() != expectedErr {
		t.Errorf("expected error %s, got %s", expectedErr, err.Error())
	}
}

func TestNewRequest(t *testing.T) {
	mockClient := &MockHttpClient{}

	mockClient.WillReturnBody("POST", "https://example.com", "mocked body")

	method := "POST"
	url := "https://example.com"
	body := bytes.NewBufferString("mocked body")

	req, err := mockClient.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	if req.Method != method {
		t.Errorf("NewRequest() method = %v, want %v", req.Method, method)
	}

	if req.URL.String() != url {
		t.Errorf("NewRequest() url = %v, want %v", req.URL.String(), url)
	}

	if body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		expectedBodyBytes, _ := io.ReadAll(bytes.NewBufferString("mocked body"))
		if !bytes.Equal(bodyBytes, expectedBodyBytes) {
			t.Errorf("NewRequest() body = %v, want %v", bodyBytes, expectedBodyBytes)
		}
	}
}

func TestDo_Success(t *testing.T) {
	mockClient := &MockHttpClient{}
	mockClient.WillReturnBody("POST", "https://example.com", "mocked response body")

	// Create a POST request that matches the mock setup
	req, _ := mockClient.NewRequest("POST", "https://example.com", bytes.NewBufferString("request body"))

	// Call the Do method
	resp, err := mockClient.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	// Validate the response
	expectedBody := "mocked response body"
	bodyBytes, _ := io.ReadAll(resp.Body)
	if !bytes.Equal(bodyBytes, []byte(expectedBody)) {
		t.Errorf("Do() body = %v, want %v", bodyBytes, expectedBody)
	}

	if resp.StatusCode != StatusOK {
		t.Errorf("Do() status code = %v, want %v", resp.StatusCode, StatusOK)
	}
}

func TestDo_Error(t *testing.T) {
	mockClient := &MockHttpClient{}

	req, _ := mockClient.NewRequest("POST", "https://example.com", bytes.NewBufferString("request body"))

	resp, err := mockClient.Do(req)
	if err == nil {
		t.Fatalf("Do() expected error, got nil")
	}

	if resp != nil {
		t.Errorf("Do() expected nil response, got %v", resp)
	}
}

func TestMockHttpClient_RegexMatching(t *testing.T) {
	client := &MockHttpClient{}

	// Define mock responses with regex patterns
	client.WillReturnBody("GET", `^/api/v1/resource/\d+$`, `{"message": "resource"}`)
	client.WillReturnBody("POST", `^/api/v1/resource/\d+/create$`, `{"message": "created"}`)
	client.WillReturnBody("GET", `^/api/v1/.*`, `{"message": "default"}`)

	tests := []struct {
		method       string
		url          string
		expectedBody string
	}{
		{"GET", "/api/v1/resource/123", `{"message": "resource"}`},
		{"POST", "/api/v1/resource/456/create", `{"message": "created"}`},
		{"GET", "/api/v1/other", `{"message": "default"}`},
		{"GET", "/api/v1/unknown", `{"message": "default"}`},
	}

	for _, test := range tests {
		resp, err := client.Do(&http.Request{
			Method: test.method,
			URL:    &url.URL{Path: test.url},
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		if string(body) != test.expectedBody {
			t.Errorf("for %s %s: expected body %s, got %s", test.method, test.url, test.expectedBody, string(body))
		}
	}
}

func TestMockHttpClient_RegexMatching2(t *testing.T) {
	client := &MockHttpClient{}

	// Define mock responses with regex patterns
	client.WillReturnBody("GET", `/api,*`, `{"message": "resource"}`)

	tests := []struct {
		method       string
		url          string
		expectedBody string
	}{
		{"GET", "/api/v1/resource/123", `{"message": "resource"}`},
	}

	for _, test := range tests {
		resp, err := client.Do(&http.Request{
			Method: test.method,
			URL:    &url.URL{Path: test.url},
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		if string(body) != test.expectedBody {
			t.Errorf("for %s %s: expected body %s, got %s", test.method, test.url, test.expectedBody, string(body))
		}
	}
}
