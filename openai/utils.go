package openai

import "strings"

type JsonDataType int

const (
	JSONObj JsonDataType = iota
	JSONArray
)

func ExtractJsonData(jsonString string, typ JsonDataType) string {
	jsonString = strings.ReplaceAll(jsonString, "\n", "")

	open, close := getBrackets(typ)

	start := strings.Index(jsonString, open)
	if start == -1 {
		return ""
	}

	end := strings.LastIndex(jsonString, close)
	if end == -1 || end <= start {
		return ""
	}

	return jsonString[start : end+1]
}

func getBrackets(typ JsonDataType) (string, string) {
	if typ == JSONArray {
		return "[", "]"
	}

	return "{", "}"
}
