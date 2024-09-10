package openai

import (
	"testing"
)

func TestExtractJsonData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		typ      JsonDataType
		expected string
	}{
		{
			name:     "Extract JSON Object",
			input:    `{"key1":"value1","key2":"value2"}`,
			typ:      JSONObj,
			expected: `{"key1":"value1","key2":"value2"}`,
		},
		{
			name:     "Extract JSON Array",
			input:    `["value1","value2","value3"]`,
			typ:      JSONArray,
			expected: `["value1","value2","value3"]`,
		},
		{
			name:     "No JSON Object",
			input:    `["value1","value2","value3"]`,
			typ:      JSONObj,
			expected: "",
		},
		{
			name:     "No JSON Array",
			input:    `{"key1":"value1","key2":"value2"}`,
			typ:      JSONArray,
			expected: "",
		},
		{
			name: "JSON Object with New Lines",
			input: `{
				"key1": "value1",
				"key2": "value2"
			}`,
			typ:      JSONObj,
			expected: `{"key1":"value1","key2":"value2"}`,
		},
		{name: "JSON Object with spaces",

			input:    `{"key1": "value1","key2": "value2"}`,
			typ:      JSONObj,
			expected: `{"key1":"value1","key2":"value2"}`,
		},
		{
			name: "JSON Array with New Lines",
			input: `[
				"value1",
				"value2",
				"value3"
			]`,
			typ:      JSONArray,
			expected: `["value1","value2","value3"]`,
		},
		{
			name:     "Empty Input",
			input:    ``,
			typ:      JSONObj,
			expected: "",
		},
		{
			name:     "Incorrectly Nested JSON should fail",
			input:    `{"key1":"value1", "key2":["value2","value3"]`,
			typ:      JSONObj,
			expected: "",
		},
		{
			name:     "Incorrectly Nested JSON should pass",
			input:    `{"key1":"value1", "key2":["value2","value3"]`,
			typ:      JSONArray,
			expected: `["value2","value3"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractJsonData(tt.input, tt.typ)
			if got != tt.expected {
				t.Errorf("ExtractJsonData() = %v, want %v", got, tt.expected)
			}
		})
	}
}
