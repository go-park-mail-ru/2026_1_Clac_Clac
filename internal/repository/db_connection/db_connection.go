package dbConnection

import (
	"sync"
	"time"

	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/google/uuid"
)

type Session struct {
	UserID    uuid.UUID
	ExpiresAt time.Time
}

type MapDatabases struct {
	UsersDB    map[uuid.UUID]models.User
	SessionsDB map[string]Session

	MutexUsers    sync.RWMutex
	MutexBoards   sync.Mutex
	MutexSessions sync.Mutex
}

func NewMapDatabse() *MapDatabases {
	return &MapDatabases{
		UsersDB:    make(map[uuid.UUID]models.User),
		SessionsDB: make(map[string]Session),
	}
}
