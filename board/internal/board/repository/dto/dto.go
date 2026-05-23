package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/common"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	"github.com/google/uuid"
)

type BoardEntry struct {
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

type MemberEntry struct {
	Link uuid.UUID
	Role rbac.Role
}

type InviteEntry struct {
	InviteLink  uuid.UUID
	BoardLink   uuid.UUID
	UserLink    *uuid.UUID
	DefaultRole rbac.Role
	ExpireTime  *time.Time
	Status      common.InviteStatus
	CreatedAt   time.Time
}

type NewInviteInfo struct {
	BoardLink   uuid.UUID
	UserLink    *uuid.UUID
	DefaultRole rbac.Role
	ExpireTime  *time.Time
}
