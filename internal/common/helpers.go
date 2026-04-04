package common

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
