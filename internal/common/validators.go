package common

import "errors"

var (
	ErrorIncorrectSymbol          = errors.New("allowed only a-z, A-Z, 0-9, and /?!@")
	ErrorIncorrectLengthName      = errors.New("name must contain maximum 128 symbols")
	ErrorIncorrectValueCountTasks = errors.New("number task must be bettwen 0 and 100")
	ErrorIncorrectColor           = errors.New("color is incoorect, can be white, grey, red, orange, blue, green, purple, pink")
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
		return ErrorIncorrectLengthName
	}
	return nil
}

func ValidateNumberInfo(info int, maxValue int, minValue int) error {
	if info > maxValue || info < minValue {
		return ErrorIncorrectValueCountTasks
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
