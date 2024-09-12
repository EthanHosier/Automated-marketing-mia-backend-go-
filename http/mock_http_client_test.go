package http_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/ethanhosier/mia-backend-go/http"
)

func TestMockHttpClient_Get_Success(t *testing.T) {
	client := &http.MockHttpClient{}
	mockURL := "http://example.com"
	mockBody := `{"message": "success"}`
	client.WillReturnBody(mockURL, mockBody)

	resp, err := client.Get(mockURL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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
	client := &http.MockHttpClient{}
	mockURL := "http://example.com"

	_, err := client.Get(mockURL)

	if err == nil {
		t.Fatalf("expected an error, got none")
	}

	expectedError := fmt.Sprintf("no mock found for %s", mockURL)
	if err.Error() != expectedError {
		t.Errorf("expected error %s, got %s", expectedError, err.Error())
	}
}

func TestMockHttpClient_Get_MultipleURLs(t *testing.T) {
	client := &http.MockHttpClient{}
	client.WillReturnBody("http://example.com/1", `{"message": "success1"}`)
	client.WillReturnBody("http://example.com/2", `{"message": "success2"}`)

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
	client := &http.MockHttpClient{}
	mockURL := "http://example.com/empty"
	client.WillReturnBody(mockURL, "")

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
	client := &http.MockHttpClient{}
	mockURL := "http://example.com"
	client.WillReturnBody(mockURL, `{"message": "old"}`)
	client.WillReturnBody(mockURL, `{"message": "new"}`)

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
