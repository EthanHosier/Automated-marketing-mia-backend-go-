package utils

import (
	"math"
	"strings"
)

type BracketType int

const (
	SquareBracket BracketType = iota // Starts from 0
	CurlyBracket                     // Increments to 1
)

func Round2Dec(val float64) float64 {
	return math.Round(val*100) / 100
}

func ExtractJsonObj(jsonString string, b BracketType) string {
	open, close := getBrackets(b)

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

func getBrackets(b BracketType) (string, string) {
	if b == SquareBracket {
		return "[", "]"
	}

	return "{", "}"
}

func RemoveDuplicates(input []string) []string {
	seen := make(map[string]struct{})
	var result []string

	for _, str := range input {
		if _, exists := seen[str]; !exists {
			seen[str] = struct{}{}
			result = append(result, str)
		}
	}

	return result
}
