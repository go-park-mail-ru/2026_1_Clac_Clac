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

type NewCard struct {
	LinkCard     uuid.UUID
	LinkAuthor   uuid.UUID
	Title        string
	Description  string
	LinkExecutor *uuid.UUID
	DataDeadLine *time.Time
	DataStart    *time.Time
	LinkSection  uuid.UUID
}

type CommentInfo struct {
	Link       uuid.UUID
	ParentLink *uuid.UUID
	AuthorLink uuid.UUID
	Text       string
	CreatedAt  time.Time
}

type CreateCommentInfo struct {
	CommentLink uuid.UUID
	CardLink    uuid.UUID
	ParentLink  *uuid.UUID
	AuthorLink  uuid.UUID
	Text        string
}

type UpdateCommentInfo struct {
	CommentLink uuid.UUID
	Text        string
}

type CreateSubtaskInfo struct {
	SubtaskLink uuid.UUID
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

type CreateAttachment struct {
	AttachmentLink uuid.UUID
	TaskLink       uuid.UUID
	Key            string
	Name           string
}

type UploadAttachment struct {
	Data        io.Reader
	FilePath    string
	ContentType string
}

type DeleteAttachmentS3 struct {
}

type UpdateStatusTask struct {
	TaskLink uuid.UUID
	Status   bool
}

type UpdateTimeLine struct {
	TaskLink uuid.UUID
	DeadLine *time.Time
	Start    *time.Time
}

type UpdateCardPoints struct {
	CardLink uuid.UUID
	Points   *int
}
