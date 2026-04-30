package common

import (
	"errors"

	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/appeal/v1"
)

var (
	ErrUnexpectedStatus = errors.New("unexpected status")
)

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

func ParseProtoStatus(pbStatus pb.Status) (Status, error) {
	switch pbStatus {
	case pb.Status_STATUS_OPEN:
		return Statuses.Open, nil
	case pb.Status_STATUS_IN_WORK:
		return Statuses.InWork, nil
	case pb.Status_STATUS_CLOSE:
		return Statuses.Close, nil
	}

	return "", ErrUnexpectedStatus
}

func ToProtoStatus(status Status) pb.Status {
	switch status {
	case Statuses.Open:
		return pb.Status_STATUS_OPEN
	case Statuses.InWork:
		return pb.Status_STATUS_IN_WORK
	case Statuses.Close:
		return pb.Status_STATUS_CLOSE
	}

	return pb.Status_STATUS_UNSPECIFIED
}
