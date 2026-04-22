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

type NewCard struct {
	LinkAuthor   uuid.UUID
	Title        string
	Description  string
	LinkExecuter *uuid.UUID
	DataDeadLine *time.Time
	LinkSection  uuid.UUID
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
