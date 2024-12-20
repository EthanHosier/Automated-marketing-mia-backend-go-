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

func RemoveElements[T comparable](slice1, slice2 []T) []T {
	toRemove := make(map[T]bool)
	for _, v := range slice2 {
		toRemove[v] = true
	}

	result := make([]T, 0, len(slice1))

	for _, v := range slice1 {
		if !toRemove[v] {
			result = append(result, v)
		}
	}

	return result
}

func GetKeysFromMap[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m)) // Create a slice with the length of the map for efficiency
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
