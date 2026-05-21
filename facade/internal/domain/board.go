package domain

import "github.com/google/uuid"

type BoardInfo struct {
	Link        uuid.UUID
	Name        string
	Description string
	Background  string
}

type GetBoardRequest struct {
	UserLink  uuid.UUID
	BoardLink uuid.UUID
}

type CreateBoardRequest struct {
	UserLink    uuid.UUID
	Name        string
	Description string
	Background  string
}

type UpdateBoardRequest struct {
	UserLink    uuid.UUID
	BoardLink   uuid.UUID
	Name        string
	Description string
	Background  string
}

type UploadBackgroundRequest struct {
	UserLink  uuid.UUID
	BoardLink uuid.UUID
	Image     []byte
	Filename  string
}

type UploadBackgroundResponse struct {
	BackgroundKey string
}

type MemberInfo struct {
	Link        uuid.UUID
	Role        string
	AvatarUrl   string
	Description string
	DisplayName string
	Email       string
}

type GetMembersRequest struct {
	UserLink  uuid.UUID
	BoardLink uuid.UUID
}

type GetMembersResponse struct {
	Members []MemberInfo
}

type CreateInviteRequest struct {
	UserLink       uuid.UUID
	BoardLink      uuid.UUID
	TargetUserLink *uuid.UUID
	DefaultRole    string
	ExpireSeconds  int64
}

type CreateInviteResponse struct {
	InviteLink     string
	BoardLink      string
	TargetUserLink *string
	DefaultRole    string
	Status         string
	ExpireAt       *int64
	CreatedAt      int64
}

type AcceptInviteRequest struct {
	InviteLink string
	UserLink   uuid.UUID
}

type CloseInviteRequest struct {
	UserLink   uuid.UUID
	InviteLink string
}

type UpdateMemberRoleRequest struct {
	UserLink       uuid.UUID
	BoardLink      uuid.UUID
	TargetUserLink uuid.UUID
	NewRole        string
}

type RemoveMemberRequest struct {
	UserLink       uuid.UUID
	BoardLink      uuid.UUID
	TargetUserLink uuid.UUID
}

type InviteInfo struct {
	InviteLink     string
	BoardLink      string
	TargetUserLink *string
	DefaultRole    string
	Status         string
	ExpireAt       *int64
	CreatedAt      int64
}

type GetActiveInvitesResponse struct {
	Invites []InviteInfo
}
