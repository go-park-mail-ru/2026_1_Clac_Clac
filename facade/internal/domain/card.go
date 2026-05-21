package domain

import (
	"io"
	"time"

	"github.com/google/uuid"
)

type CardFullInfo struct {
	CardLink     uuid.UUID
	ExecutorLink *uuid.UUID
	Title        string
	Description  string
	Deadline     *time.Time
	Start        *time.Time
	Status       bool
	Subtasks     []SubtaskInfo
	Position     int
	Attachments  []AttachmentInfo
}

type AttachmentInfo struct {
	AttachmentLink uuid.UUID
	DisplayName    string
	Path           string
	Position       int
}

type GetCardRequest struct {
	UserLink uuid.UUID
	CardLink uuid.UUID
}

type DeleteCardRequest struct {
	UserLink uuid.UUID
	CardLink uuid.UUID
}

type UpdateCardRequest struct {
	UserLink     uuid.UUID
	CardLink     uuid.UUID
	ExecutorLink *uuid.UUID
	Title        string
	Description  string
	Deadline     *time.Time
	Start        *time.Time
}

type ReorderCardsRequest struct {
	UserLink    uuid.UUID
	CardLink    uuid.UUID
	SectionLink uuid.UUID
	Position    int
}

type CreateCardRequest struct {
	UserLink     uuid.UUID
	SectionLink  uuid.UUID
	ExecutorLink *uuid.UUID
	Title        string
}

type CreateCardResponse struct {
	CardLink    uuid.UUID
	SectionLink uuid.UUID
	Position    int
}

type GetCommentsRequest struct {
	UserLink uuid.UUID
	CardLink uuid.UUID
}

type GetCommentsResponse struct {
	CommentsInfo []CommentInfo
}

type CommentInfo struct {
	CommentLink uuid.UUID
	ParentLink  *uuid.UUID
	AuthorLink  uuid.UUID
	Text        string
	CreatedAt   time.Time
}

type CreateCommentRequest struct {
	UserLink   uuid.UUID
	CardLink   uuid.UUID
	ParentLink *uuid.UUID
	Text       string
}

type CreateCommentResponse struct {
	CommentLink uuid.UUID
}

type DeleteCommentRequest struct {
	UserLink    uuid.UUID
	CommentLink uuid.UUID
}

type UpdateCommentRequest struct {
	UserLink    uuid.UUID
	CommentLink uuid.UUID
	Text        string
}

type CreateSubtaskRequest struct {
	UserLink    uuid.UUID
	CardLink    uuid.UUID
	Description string
}

type CreateSubtaskResponse struct {
	SubtaskLink uuid.UUID
	Description string
}

type UpdateSubtaskRequest struct {
	UserLink    uuid.UUID
	SubtaskLink uuid.UUID
	IsDone      bool
	Description string
}

type DeleteSubtaskRequest struct {
	UserLink    uuid.UUID
	SubtaskLink uuid.UUID
}

type CreateAttachmentRequest struct {
	UserLink   uuid.UUID
	TaskLink   uuid.UUID
	Attachment io.Reader
	Filename   string
}

type DeleteAttachmentRequest struct {
	UserLink       uuid.UUID
	AttachmentLink uuid.UUID
}

type NewStatusTask struct {
	UserLink uuid.UUID
	CardLink uuid.UUID
	Status   bool
}

type NewTimeLine struct {
	UserLink uuid.UUID
	CardLink uuid.UUID
	DeadLine time.Time
	Start    time.Time
}
