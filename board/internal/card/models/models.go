package models

import "github.com/google/uuid"

type SubtaskInfo struct {
	SubtaskLink uuid.UUID
	Description string
	IsDone      bool
	Position    int
}
