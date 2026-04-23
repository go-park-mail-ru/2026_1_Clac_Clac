package dto

import (
	"time"

	"github.com/google/uuid"
)

type InfoCard struct {
	Description  string
	Title        string
	NameExecuter *string
	DataDeadLine *time.Time
}

type UpdatingCardDetails struct {
	LinkCard     uuid.UUID
	Title        string
	Description  string
	LinkExecuter *uuid.UUID
	DataDeadLine *time.Time
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
	LinkExecuter *uuid.UUID
	DataDeadLine *time.Time
	LinkSection  uuid.UUID
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
	Text        string
}
