package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/common"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	"github.com/google/uuid"
)

type BoardInfo struct {
	Link        uuid.UUID
	Name        string
	Description string
	Background  string
	CreatedAt   time.Time
}

type NewBoardInfo struct {
	Name        string
	Description string
	Background  string
}

type UpdateBoardInfo struct {
	Link        uuid.UUID
	Name        string
	Description string
	Background  string
}

type InviteInfo struct {
	InviteLink   uuid.UUID
	BoardLink    uuid.UUID
	TargetUser   *uuid.UUID
	DefaultRole  rbac.Role
	Status       common.InviteStatus
	ExpireAt     *time.Time
	CreatedAt    time.Time
}

type NewInviteInfo struct {
	BoardLink    uuid.UUID
	UserLink     *uuid.UUID
	DefaultRole  rbac.Role
	ExpireTime   *time.Time
}
