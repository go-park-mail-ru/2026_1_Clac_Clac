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

func CheckSectionNameLength(name string, maxLen int) bool {
	return len(name) <= maxLen
}

func CheckMaxTasks(tasks int, maxValue int, minValue int) bool {
	return (tasks > minValue) && (maxValue > tasks)
}

func CheckColor(color string) bool {
	_, ok := vaildColor[color]
	return ok
}
