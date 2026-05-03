package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/models"
	"github.com/google/uuid"
)

type InfoCard struct {
	Description   string
	Title        string
	ExecutorLink *uuid.UUID
	DataDeadLine  *time.Time
	Subtasks     []models.SubtaskInfo
	Position    int
}

type NewCard struct {
	LinkAuthor   uuid.UUID
	Title        string
	Description  string
	LinkExecutor *uuid.UUID
	DataDeadLine *time.Time
	LinkSection  uuid.UUID
}

type UpdatingCardDetails struct {
	LinkCard     uuid.UUID
	Title        string
	Description  string
	LinkExecutor *uuid.UUID
	DataDeadLine *time.Time
}

type PlaceCard struct {
	LinkCard    uuid.UUID
	LinkSection uuid.UUID
	Position    int
}

type CommentInfo struct {
	Link       uuid.UUID
	ParentLink *uuid.UUID
	AuthorLink uuid.UUID
	Text       string
}

type CreateCommentInfo struct {
	CardLink   uuid.UUID
	ParentLink *uuid.UUID
	AuthorLink uuid.UUID
	Text       string
}

type UpdateCommentInfo struct {
	CommentLink uuid.UUID
	UserLink    uuid.UUID
	Text        string
}

type CreateSubtaskInfo struct {
	TaskLink    uuid.UUID
	Description string
}

type DeleteSubtask struct {
	SubTaskLink uuid.UUID
}

type UpdateSubtask struct {
	SubTaskLink uuid.UUID
	Description string
	IsDone      bool
}
