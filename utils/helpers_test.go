package utils

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func mockFuncSuccessAfterFailures(failuresBeforeSuccess int) RetryFunc[int] {
	attempt := 0
	return func() (int, error) {
		attempt++
		if attempt <= failuresBeforeSuccess {
			return 0, errors.New("mock failure")
		}
		return 42, nil
	}
}

func mockFuncAlwaysFail() RetryFunc[int] {
	return func() (int, error) {
		return 0, errors.New("always fails")
	}
}

func TestRetry_SuccessAfterFailures(t *testing.T) {
	retries := 5
	result, err := Retry(retries, mockFuncSuccessAfterFailures(3))

	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if result != 42 {
		t.Fatalf("expected result 42, got: %d", result)
	}
}

func TestRetry_AlwaysFails(t *testing.T) {
	retries := 3
	result, err := Retry(retries, mockFuncAlwaysFail())

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if result != 0 {
		t.Fatalf("expected result 0 on failure, got: %d", result)
	}
}

func TestDoAsync_Success(t *testing.T) {
	task := DoAsync(func() (int, error) {
		time.Sleep(100 * time.Millisecond) // Simulating some work
		return 42, nil
	})

	result, err := GetAsync(task)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != 42 {
		t.Fatalf("expected result 42, got: %d", result)
	}
}

// Test DoAsync with a function that returns an error
func TestDoAsync_Error(t *testing.T) {
	task := DoAsync(func() (int, error) {
		time.Sleep(100 * time.Millisecond) // Simulating some work
		return 0, errors.New("some error occurred")
	})

	result, err := GetAsync(task)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
	if result != 0 {
		t.Fatalf("expected result 0 on error, got: %d", result)
	}
}

// Test DoAsync with a string type
func TestDoAsync_StringSuccess(t *testing.T) {
	task := DoAsync(func() (string, error) {
		time.Sleep(100 * time.Millisecond)
		return "hello", nil
	})

	result, err := GetAsync(task)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != "hello" {
		t.Fatalf("expected result 'hello', got: %s", result)
	}
}

func TestDoAsync_Goroutine(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	task := DoAsync(func() (int, error) {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond)
		return 42, nil
	})

	select {
	case <-task.ch:
		t.Fatal("expected task to not complete immediately")
	case <-time.After(50 * time.Millisecond):
		t.Log("function is running asynchronously, not completed yet")
	}

	wg.Wait()

	result, err := GetAsync(task)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != 42 {
		t.Fatalf("expected result 42, got: %d", result)
	}
}

func mockFuncListSuccess(item int) (string, error) {
	if item%2 == 0 {
		return "even", nil
	}
	return "odd", nil
}

func mockFuncListFail(item int) (string, error) {
	return "", errors.New("mock failure")
}

func TestDoAsyncList_Success(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	tasks := DoAsyncList(items, mockFuncListSuccess)

	results, err := GetAsyncList(tasks)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expected := []string{"odd", "even", "odd", "even", "odd"}
	for i, result := range results {
		if result != expected[i] {
			t.Fatalf("expected result %s, got: %s", expected[i], result)
		}
	}
}

func TestDoAsyncList_Failures(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	tasks := DoAsyncList(items, mockFuncListFail)

	_, err := GetAsyncList(tasks)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestDoAsyncList_EmptyList(t *testing.T) {
	items := []int{}
	tasks := DoAsyncList(items, mockFuncListSuccess)

	results, err := GetAsyncList(tasks)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("expected empty result list, got: %v", results)
	}
}

func TestGetAsyncList_WithErrors(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	tasks := DoAsyncList(items, func(item int) (string, error) {
		if item%2 == 0 {
			return "", errors.New("error for even item")
		}
		return "odd", nil
	})

	results, err := GetAsyncList(tasks)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if results != nil {
		t.Fatalf("expected nil results on error, got: %v", results)
	}
}
