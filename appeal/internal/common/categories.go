package common

type Category string

var Categories = struct {
	Bug       Category
	Proposal  Category
	Complaint Category
}{
	Bug:       "bug",
	Proposal:  "proposal",
	Complaint: "complaint",
}
