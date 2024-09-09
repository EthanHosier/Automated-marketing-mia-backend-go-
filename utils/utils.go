package utils

import (
	"errors"
	"strconv"
	"unicode"
)

func FirstNumberInString(s string) (int, error) {
	num := ""
	for _, char := range s {
		if unicode.IsDigit(char) {
			num += string(char)
		} else if len(num) > 0 {
			break
		}
	}

	if num == "" {
		return 0, errors.New("no number found in the string")
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
