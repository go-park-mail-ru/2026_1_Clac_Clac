package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateBoardRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Background  string `json:"background"`
}

type DeleteBoardRequest struct {
	Link uuid.UUID `json:"link"`
}

type UpdateBoardRequest struct {
	Link        uuid.UUID `json:"link"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Background  string    `json:"background"`
}

type BoardInfoResponse struct {
	Link        uuid.UUID `json:"link"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Background  string    `json:"background"`
	CreatedAt   time.Time `json:"created_at"`
}

type BackgroundUpdateResponse struct {
	BackgroundURL string `json:"background_url"`
}
