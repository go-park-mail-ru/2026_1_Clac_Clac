package repository

import (
	"context"
	"errors"
	"sync"
	"time"

	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/google/uuid"
)

var (
	ErrorExistingUser       = errors.New("user with this email alreday exists")
	ErrorNonexistentUser    = errors.New("user with this email not exist")
	ErrorDetectingCollision = errors.New("session collision detected")

	ErrorNotExistingSession = errors.New("session not found or expired")
	ErrorSeesionExpired     = errors.New("time life session expired")
)

func NewMapDB() *MapDatabases {
	return &MapDatabases{
		database: make(map[string]models.User),
		sessions: make(map[string]Session),
	}
}

type Session struct {
	UserID    uuid.UUID
	ExpiresAt time.Time
}

type MapDatabases struct {
	database      map[string]models.User
	sessions      map[string]Session
	mutexUsers    sync.RWMutex
	mutexSessions sync.RWMutex
}

func (mp *MapDatabases) AddUser(ctx context.Context, user models.User) error {
	mp.mutexUsers.Lock()
	defer mp.mutexUsers.Unlock()

	if _, exist := mp.database[user.Email]; exist {
		return ErrorExistingUser
	}
	mp.database[user.Email] = user

	return nil
}

func (mp *MapDatabases) AddSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	mp.mutexUsers.Lock()
	defer mp.mutexUsers.Unlock()

	if _, exist := mp.sessions[sessionID]; exist {
		return ErrorDetectingCollision
	}

	mp.sessions[sessionID] = Session{
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return nil
}

func (mp *MapDatabases) GetUserIDBySession(ctx context.Context, sessionID string) (uuid.UUID, error) {
	mp.mutexUsers.RLock()
	session, exist := mp.sessions[sessionID]
	mp.mutexUsers.RUnlock()

	if !exist {
		return uuid.Nil, ErrorNotExistingSession
	}

	if time.Now().After(session.ExpiresAt) {
		mp.mutexUsers.Lock()
		delete(mp.sessions, sessionID)
		mp.mutexUsers.Unlock()
		return uuid.Nil, ErrorSeesionExpired
	}

	return session.UserID, nil
}

func (mp *MapDatabases) DeleteSession(ctx context.Context, sessionID string) error {
	_, exist := mp.sessions[sessionID]
	if !exist {
		return ErrorNotExistingSession
	}

	mp.mutexUsers.Lock()
	delete(mp.sessions, sessionID)
	mp.mutexUsers.Unlock()
	return nil
}

func (mp *MapDatabases) GetUser(ctx context.Context, email string) (models.User, error) {
	mp.mutexUsers.RLock()
	defer mp.mutexUsers.RUnlock()

	if user, exist := mp.database[email]; exist {
		return user, nil
	}

	return models.User{}, ErrorNonexistentUser
}
