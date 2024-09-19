package utils

import (
	"fmt"
	"strconv"
	"unicode"
)

func FirstNumberInString(s string) (int, error) {
	num := ""
	neg := false

	for i, char := range s {
		if char == '-' && i+1 < len(s) && unicode.IsDigit(rune(s[i+1])) {
			neg = true
			continue
		}
		if unicode.IsDigit(char) {
			num += string(char)
		} else if len(num) > 0 {
			break
		}
	}

	if num == "" {
		return 0, fmt.Errorf("no number found in the string %s", s)
	}

	if neg {
		num = "-" + num
	}

	number, err := strconv.Atoi(num)
	if err != nil {
		return 0, err
	}

	return number, nil
}

func FirstNChars(text string, n int) string {
	runes := []rune(text)
	if len(runes) < n {
		n = len(runes)
	}
	return string(runes[:n])
}

func Flatten[T any](sliceOfSlices [][]T) []T {
	var flatSlice []T
	for _, innerSlice := range sliceOfSlices {
		flatSlice = append(flatSlice, innerSlice...)
	}
	if flatSlice == nil {
		return []T{}
	}
	return flatSlice
}

func RemoveDuplicates[T comparable](slice []T) []T {
	encountered := map[T]bool{}
	result := []T{}

	for _, v := range slice {
		if !encountered[v] {
			encountered[v] = true
			result = append(result, v)
		}
	}
	return result
}
