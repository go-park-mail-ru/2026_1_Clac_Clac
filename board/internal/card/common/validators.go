package common

var vaildColor = map[string]struct{}{
	"white":  {},
	"grey":   {},
	"red":    {},
	"orange": {},
	"blue":   {},
	"green":  {},
	"purple": {},
	"pink":   {},
}

func CheckCardNameLength(name string, maxLen int) bool {
	return len(name) <= maxLen
}

func CheckCardDescriptionLength(description string, maxLen int) bool {
	return len(description) <= maxLen
}

func CheckColor(color string) bool {
	_, ok := vaildColor[color]
	return ok
}
