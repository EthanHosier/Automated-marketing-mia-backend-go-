package utils

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestFirstNumberInString(t *testing.T) {
	tests := []struct {
		input          string
		expectedNumber int
		expectedError  bool
	}{
		{"abc123def", 123, false},               // Positive number in the string
		{"-456xyz", -456, false},                // Negative number at the start
		{"abc-789", -789, false},                // Negative number in the middle
		{"no numbers", 0, true},                 // No numbers in the string
		{"--123", -123, false},                  // Invalid number (extra '-')
		{"123abc-456", 123, false},              // First positive number
		{"", 0, true},                           // Empty string
		{"something-123again-456", -123, false}, // First negative number
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := FirstNumberInString(test.input)

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedNumber, result)
			}
		})
	}
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		input          [][]int
		expectedOutput []int
	}{
		{ // Test with multiple slices containing integers
			input:          [][]int{{1, 2, 3}, {4, 5}, {6}},
			expectedOutput: []int{1, 2, 3, 4, 5, 6},
		},
		{ // Test with an empty slice of slices
			input:          [][]int{},
			expectedOutput: []int{},
		},
		{ // Test with a single empty slice
			input:          [][]int{{}},
			expectedOutput: []int{},
		},
		{ // Test with slices containing mixed elements
			input:          [][]int{{7}, {8, 9}, {10, 11, 12}},
			expectedOutput: []int{7, 8, 9, 10, 11, 12},
		},
		{ // Test with slices of different lengths
			input:          [][]int{{1, 2}, {3}, {4, 5, 6}},
			expectedOutput: []int{1, 2, 3, 4, 5, 6},
		},
	}

	for _, test := range tests {
		t.Run("FlattenTest", func(t *testing.T) {
			result := Flatten(test.input)
			assert.Equal(t, test.expectedOutput, result)
		})
	}
}

func TestRemoveDuplicatesInt(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "No duplicates",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "With duplicates",
			input:    []int{1, 2, 2, 3, 4, 5, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "All duplicates",
			input:    []int{1, 1, 1, 1, 1},
			expected: []int{1},
		},
		{
			name:     "Empty slice",
			input:    []int{},
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveDuplicates(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRemoveElementsInt(t *testing.T) {
	slice1 := []int{1, 2, 3, 4, 5}
	slice2 := []int{2, 4}
	expected := []int{1, 3, 5}

	result := RemoveElements(slice1, slice2)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestRemoveElementsString(t *testing.T) {
	slice1 := []string{"apple", "banana", "cherry", "banana"}
	slice2 := []string{"banana"}
	expected := []string{"apple", "cherry"}

	result := RemoveElements(slice1, slice2)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestRemoveElementsNoMatch(t *testing.T) {
	slice1 := []int{1, 2, 3}
	slice2 := []int{4, 5}
	expected := slice1 // No elements should be removed

	result := RemoveElements(slice1, slice2)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestRemoveElementsAllMatch(t *testing.T) {
	slice1 := []int{1, 2, 3}
	slice2 := []int{1, 2, 3}
	expected := []int{} // All elements should be removed

	result := RemoveElements(slice1, slice2)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestRemoveElementsEmptySlices(t *testing.T) {
	slice1 := []int{}
	slice2 := []int{}
	expected := []int{} // No elements in both slices

	result := RemoveElements(slice1, slice2)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestRemoveElementsPartialMatch(t *testing.T) {
	slice1 := []string{"red", "green", "blue"}
	slice2 := []string{"green", "purple"}
	expected := []string{"red", "blue"} // "green" should be removed

	result := RemoveElements(slice1, slice2)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestGetKeysFromMapStringInt(t *testing.T) {
	myMap := map[string]int{
		"apple":  1,
		"banana": 2,
		"cherry": 3,
	}

	expected := []string{"apple", "banana", "cherry"}
	result := GetKeysFromMap(myMap)

	// Since map iteration order is random, we can check for set equality
	if !equalUnordered(expected, result) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestGetKeysFromMapIntString(t *testing.T) {
	myMap := map[int]string{
		1: "one",
		2: "two",
		3: "three",
	}

	expected := []int{1, 2, 3}
	result := GetKeysFromMap(myMap)

	// Since map iteration order is random, we can check for set equality
	if !equalUnordered(expected, result) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestGetKeysFromEmptyMap(t *testing.T) {
	myMap := map[string]int{}

	expected := []string{}
	result := GetKeysFromMap(myMap)

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

// Helper function to compare two slices, ignoring order
func equalUnordered[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	// Create a map to count occurrences of each element in `a`
	counts := make(map[T]int)
	for _, v := range a {
		counts[v]++
	}

	// Decrement the counts based on elements in `b`
	for _, v := range b {
		counts[v]--
		if counts[v] < 0 {
			return false
		}
	}

	// Check if all counts are 0, meaning all elements matched
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}

	return true
}
