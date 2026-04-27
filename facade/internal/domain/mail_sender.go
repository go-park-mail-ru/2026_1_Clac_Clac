package domain

import "github.com/google/uuid"

type RecoveryCode struct {
	UserLink uuid.UUID
	Email    string
}

type RecoveryCodeCheck struct {
	Code string
}

type ResetToken struct {
	Token string
}
