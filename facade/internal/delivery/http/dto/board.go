package dto

import "github.com/google/uuid"

type BoardInfo struct {
	Link        uuid.UUID `json:"link"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Background  string    `json:"background"`
}

type GetBoardRequest struct {
	UserLink  uuid.UUID `json:"user_link"`
	BoardLink uuid.UUID `json:"board_link"`
}

type CreateBoardRequest struct {
	UserLink    uuid.UUID `json:"user_link"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Background  string    `json:"background"`
}

type UpdateBoardRequest struct {
	UserLink    uuid.UUID `json:"user_link"`
	BoardLink   uuid.UUID `json:"board_link"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Background  string    `json:"background"`
}

type UploadBackgroundRequest struct {
	UserLink  uuid.UUID `json:"user_link"`
	BoardLink uuid.UUID `json:"board_link"`
	Filename  string    `json:"filename"`
}

type UploadBackgroundResponse struct {
	BackgroundKey string `json:"background_key"`
}

type GetMembersRequest struct {
	UserLink  uuid.UUID `json:"user_link"`
	BoardLink uuid.UUID `json:"board_link"`
}

type GetMembersResponse struct {
	UserLinks []uuid.UUID `json:"user_links"`
}
