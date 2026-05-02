package domain

import "github.com/google/uuid"

type BoardInfo struct {
	Link        uuid.UUID
	Name        string
	Description string
	Background  string
}

type GetBoardRequest struct {
	UserLink  uuid.UUID
	BoardLink uuid.UUID
}

type CreateBoardRequest struct {
	UserLink    uuid.UUID
	Name        string
	Description string
	Background  string
}

type UpdateBoardRequest struct {
	UserLink    uuid.UUID
	BoardLink   uuid.UUID
	Name        string
	Description string
	Background  string
}

type UploadBackgroundRequest struct {
	UserLink  uuid.UUID
	BoardLink uuid.UUID
	Image     []byte
	Filename  string
}

type UploadBackgroundResponse struct {
	BackgroundKey string
}

type GetMembersRequest struct {
	UserLink  uuid.UUID
	BoardLink uuid.UUID
}

type GetMembersResponse struct {
	UserLinks []uuid.UUID
}
