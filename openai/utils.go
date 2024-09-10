package openai

import (
	"encoding/json"
	"strings"
)

type JsonDataType int

const (
	JSONObj JsonDataType = iota
	JSONArray
)

// ExtractJsonData extracts a JSON object or array from a string
func ExtractJsonData(jsonString string, typ JsonDataType) string {
	// Remove new lines and excessive spaces
	jsonString = strings.Join(strings.Fields(jsonString), "")

	open, close := getBrackets(typ)

	// Find the start and end of the JSON data
	start := strings.Index(jsonString, open)
	if start == -1 {
		return ""
	}

	end := strings.LastIndex(jsonString, close)
	if end == -1 || end <= start {
		return ""
	}

	// Extract the potential JSON data
	result := jsonString[start : end+1]

	// Validate JSON structure
	if isValidJson(result, typ) {
		return result
	}
	return ""
}

func getBrackets(typ JsonDataType) (string, string) {
	if typ == JSONArray {
		return "[", "]"
	}
	return "{", "}"
}

// isValidJson checks if the data is valid JSON and of the expected type
func isValidJson(data string, typ JsonDataType) bool {
	var js json.RawMessage
	if err := json.Unmarshal([]byte(data), &js); err != nil {
		return false
	}

	// Check if the JSON matches the expected type
	if typ == JSONObj {
		return isJSONObject(data)
	} else if typ == JSONArray {
		return isJSONArray(data)
	}
	return false
}

func isJSONObject(data string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(data), &js) == nil
}

func isJSONArray(data string) bool {
	var js []interface{}
	return json.Unmarshal([]byte(data), &js) == nil
}
