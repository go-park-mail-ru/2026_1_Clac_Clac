package common

type Category string

var Categories = struct {
	Bug       Status
	Proposal  Status
	Complaint Status
}{
	Bug:       "bug",
	Proposal:  "proposal",
	Complaint: "complaint",
}
