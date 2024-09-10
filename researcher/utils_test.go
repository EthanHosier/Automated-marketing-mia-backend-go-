package researcher

import (
	"testing"
)

func TestSortURLsByProximity(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
		err      bool
	}{
		{
			input:    []string{"https://example.com/a/b/c", "https://example.com/a", "https://example.com/a/b"},
			expected: []string{"https://example.com/a", "https://example.com/a/b", "https://example.com/a/b/c"},
			err:      false,
		},
		{
			input:    []string{"https://example.com/a/b/c", "https://example.com/a/b/c/d", "https://example.com"},
			expected: []string{"https://example.com", "https://example.com/a/b/c", "https://example.com/a/b/c/d"},
			err:      false,
		},
	}

	for _, tt := range tests {
		got, err := sortURLsByProximity(tt.input)
		if (err != nil) != tt.err {
			t.Errorf("sortURLsByProximity(%v) error = %v, wantErr %v", tt.input, err, tt.err)
			continue
		}
		if !equalStringSlices(got, tt.expected) {
			t.Errorf("sortURLsByProximity(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{
			input:    []string{"https://example.com", "https://example.com", "https://example.org"},
			expected: []string{"https://example.com", "https://example.org"},
		},
		{
			input:    []string{"https://example.com/a", "https://example.com/a", "https://example.com/b"},
			expected: []string{"https://example.com/a", "https://example.com/b"},
		},
		{
			input:    []string{"https://example.com"},
			expected: []string{"https://example.com"},
		},
		{
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		got := removeDuplicates(tt.input)
		if !equalStringSlices(got, tt.expected) {
			t.Errorf("removeDuplicates(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

// Helper function to compare two slices of strings
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
