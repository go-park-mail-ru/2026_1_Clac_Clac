package dto

import (
	"io"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/models"
	"github.com/google/uuid"
)

type InfoCard struct {
	Description  string
	Title        string
	ExecutorLink *uuid.UUID
	DataDeadLine *time.Time
	DataStart    *time.Time
	Status       bool
	Subtasks     []models.SubtaskInfo
	Position     int
	Attachments  []models.AttachmentInfo
}

type NewCard struct {
	LinkAuthor   uuid.UUID
	Title        string
	Description  string
	LinkExecutor *uuid.UUID
	DataDeadLine *time.Time
	DataStart    *time.Time
	LinkSection  uuid.UUID
}

type UpdatingCardDetails struct {
	LinkCard     uuid.UUID
	Title        string
	Description  string
	LinkExecutor *uuid.UUID
	DataDeadLine *time.Time
	DataStart    *time.Time
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
	CreatedAt  time.Time
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

type AttachmentInfo struct {
	AttachmentLink uuid.UUID
	Path           string
	Position       int
	DisplayName    string
}

type CreateAttachment struct {
	TaskLink    uuid.UUID
	UserLink    uuid.UUID
	Data        io.Reader
	ContentType string
	Extension   string
	DisplayName string
}

type DeleteAttachment struct {
	AttachmentLink uuid.UUID
	UserLink       uuid.UUID
}

type UpdateStatusTask struct {
	TaskLink uuid.UUID
	UserLink uuid.UUID
	Status   bool
}

type UpdateTimeLine struct {
	TaskLink uuid.UUID
	UserLink uuid.UUID
	DeadLine *time.Time
	Start    *time.Time
}

type UpdateCardPoints struct {
	CardLink uuid.UUID
	UserLink uuid.UUID
	Points   *int
}
