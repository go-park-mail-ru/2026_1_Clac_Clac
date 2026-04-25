package dto

import "github.com/google/uuid"

type EntityAppeal struct {
	UserLink    uuid.UUID
	Mail        string
	Category    string
	Description string
	DisplayName string
}
