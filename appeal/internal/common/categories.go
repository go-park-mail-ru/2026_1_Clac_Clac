package common

import (
	"errors"

	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/appeal/v1"
)

var (
	ErrUnexpectedCategory = errors.New("unexpected category")
)

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

func ParseProtoCategory(pbCategory pb.Category) (Category, error) {
	switch pbCategory {
	case pb.Category_CATEGORY_BUG:
		return Categories.Bug, nil
	case pb.Category_CATEGORY_PROPOSAL:
		return Categories.Proposal, nil
	case pb.Category_CATEGORY_COMPLAINT:
		return Categories.Complaint, nil
	}

	return "", ErrUnexpectedCategory
}

func ToProtoCategory(category Category) pb.Category {
	switch category {
	case Categories.Bug:
		return pb.Category_CATEGORY_BUG
	case Categories.Proposal:
		return pb.Category_CATEGORY_PROPOSAL
	case Categories.Complaint:
		return pb.Category_CATEGORY_COMPLAINT
	}

	return pb.Category_CATEGORY_UNSPECIFIED
}
