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
)

type Database interface {
	AddUser(ctx context.Context, user models.User) error
	AddSession(ctx context.Context, userID uuid.UUID, sessionID string) error
}

func NewMapDB() *MapDatabases {
	return &MapDatabases{
		database: make(map[string]models.User),
		sessions: make(map[string]uuid.UUID),
	}
}

type Session struct {
	ID        uuid.UUID
	ExpiresAt time.Time
}

type MapDatabases struct {
	database map[string]models.User
	sessions map[string]uuid.UUID
	mutex    sync.RWMutex
}

func (mp *MapDatabases) AddUser(ctx context.Context, user models.User) error {
	if _, exist := mp.database[user.Email]; exist {
		return ErrorExistingUser
	}
	mp.database[user.Email] = user

	return nil
}

func (mp *MapDatabases) AddSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	if _, exist := mp.sessions[sessionID]; exist {
		return ErrorDetectingCollision
	}

	mp.sessions[sessionID] = userID

	return nil
}

func (mp *MapDatabases) GetUser(ctx context.Context, email string) (models.User, error) {
	if userData, exist := mp.database[email]; !exist {
		return userData, nil
	}

	return models.User{}, ErrorNonexistentUser
}
