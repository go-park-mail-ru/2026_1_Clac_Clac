package db

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type ResetToken struct {
	ResetTokenID string
	UserID       uuid.UUID
	ExpiresAt    time.Time
}

type Session struct {
	SessionID string
	UserID    uuid.UUID
	ExpiresAt time.Time
}

type User struct {
	ID           uuid.UUID
	DisplayName  string
	PasswordHash string
	Email        string
	Avatar       *string
	Boards       []Board
}

type Board struct {
	ID uuid.UUID
}

type MapDatabases struct {
	UsersDB       map[uuid.UUID]User
	SessionsDB    map[string]Session
	ResetTokensDB map[string]ResetToken

	MutexUsers    sync.RWMutex
	MutexBoards   sync.Mutex
	MutexSessions sync.Mutex
	MutexTokens   sync.RWMutex
}

func NewMapDatabse() *MapDatabases {
	return &MapDatabases{
		UsersDB:       make(map[uuid.UUID]User),
		SessionsDB:    make(map[string]Session),
		ResetTokensDB: make(map[string]ResetToken),
	}
}
