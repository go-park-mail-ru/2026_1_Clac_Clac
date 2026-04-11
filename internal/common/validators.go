package common

import (
	"fmt"
)

var (
	vaildColor = map[string]struct{}{
		"white":  {},
		"grey":   {},
		"red":    {},
		"orange": {},
		"blue":   {},
		"green":  {},
		"purple": {},
		"pink":   {},
	}
)

const maxAsciiCode = 127

func CheckAsciiSymbol(strings ...string) bool {
	for _, str := range strings {
		for _, symbol := range str {
			if symbol > maxAsciiCode {
				return false
			}
		}
	}

	return true
}

func ValidateTextInfo(info string, maxLen int) error {
	if len(info) > maxLen {
		return fmt.Errorf("must contain maximum %d symbols", maxLen)
	}
	return nil
}

func ValidateNumberInfo(info int, maxValue int, minValue int) error {
	if info > maxValue || info < minValue {
		return fmt.Errorf("number must be bettwin %d and %d", minValue, maxValue)
	}

	return nil
}

func ValidateColor(color string) error {
	_, ok := vaildColor[color]
	if !ok {
		return ErrorIncorrectColor
	}

	return nil
}
