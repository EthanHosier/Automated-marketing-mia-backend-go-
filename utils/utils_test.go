package utils

import (
	"strings"
	"testing"
)

// Test when the string contains a number in the middle
func TestFirstNumberInString_WithNumber(t *testing.T) {
	result, err := FirstNumberInString("abc123def")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != 123 {
		t.Fatalf("expected result 123, got: %d", result)
	}
}

// Test when the string starts with a number
func TestFirstNumberInString_NumberAtStart(t *testing.T) {
	result, err := FirstNumberInString("456abc")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != 456 {
		t.Fatalf("expected result 456, got: %d", result)
	}
}

// Test when the string has no numbers
func TestFirstNumberInString_NoNumber(t *testing.T) {
	_, err := FirstNumberInString("abcdef")
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "no number found in the string") {
		t.Fatalf("expected error 'no number found in the string', got: %v", err)
	}
}

// Test when the string contains multiple numbers, only first should be returned
func TestFirstNumberInString_MultipleNumbers(t *testing.T) {
	result, err := FirstNumberInString("abc123def456")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != 123 {
		t.Fatalf("expected result 123, got: %d", result)
	}
}

// Test when the string contains numbers separated by letters
func TestFirstNumberInString_InterspersedNumbers(t *testing.T) {
	result, err := FirstNumberInString("abc12def34")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != 12 {
		t.Fatalf("expected result 12, got: %d", result)
	}
}

// Test when n is less than the length of the string
func TestFirstNChars_LessThanLength(t *testing.T) {
	result := FirstNChars("hello world", 5)
	expected := "hello"
	if result != expected {
		t.Fatalf("expected '%s', got '%s'", expected, result)
	}
}

// Test when n is more than the length of the string
func TestFirstNChars_MoreThanLength(t *testing.T) {
	result := FirstNChars("hello", 10)
	expected := "hello"
	if result != expected {
		t.Fatalf("expected '%s', got '%s'", expected, result)
	}
}

// Test when n is equal to the length of the string
func TestFirstNChars_EqualToLength(t *testing.T) {
	result := FirstNChars("world", 5)
	expected := "world"
	if result != expected {
		t.Fatalf("expected '%s', got '%s'", expected, result)
	}
}

// Test when n is 0
func TestFirstNChars_Zero(t *testing.T) {
	result := FirstNChars("world", 0)
	expected := ""
	if result != expected {
		t.Fatalf("expected empty string, got '%s'", result)
	}
}

// Test when the string contains non-ASCII characters
func TestFirstNChars_WithNonASCII(t *testing.T) {
	result := FirstNChars("こんにちは世界", 5) // Japanese for "Hello, World"
	expected := "こんにちは"
	if result != expected {
		t.Fatalf("expected '%s', got '%s'", expected, result)
	}
}
