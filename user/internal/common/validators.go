package common

import "fmt"

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
