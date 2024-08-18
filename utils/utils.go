package utils

import (
	"math"
	"strings"
)

func Round2Dec(val float64) float64 {
	return math.Round(val*100) / 100
}

func ExtractJsonObj(jsonString string) string {
	start := strings.Index(jsonString, "{")
	if start == -1 {
		return ""
	}

	end := strings.LastIndex(jsonString, "}")
	if end == -1 || end <= start {
		return ""
	}

	return jsonString[start : end+1]
}
