package common

type Status string

var Statuses = struct {
	Open   Status
	InWork Status
	Close  Status
}{
	Open:   "new",
	InWork: "in_progress",
	Close:  "closed",
}
