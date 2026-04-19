package dto

import (
	"time"

	"github.com/google/uuid"
)

type BoardEntry struct {
	Link        uuid.UUID
	Name        string
	Description string
	Background  string
	CreatedAt   time.Time
}

type NewBoardInfo struct {
	Name        string
	Description string
	Background  string
}

type UpdateBoardInfo struct {
	Link        uuid.UUID
	Name        string
	Description string
	Background  string
}
