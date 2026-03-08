package auth

import (
	"context"
	"time"

	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	repository "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository"
	dbConnection "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/db_connection"
	"github.com/google/uuid"
)

type AuthRepository struct {
	database *dbConnection.MapDatabases
}

func NewAuthRepository(db *dbConnection.MapDatabases) *AuthRepository {
	return &AuthRepository{
		database: db,
	}
}

func (ar *AuthRepository) AddUser(ctx context.Context, user models.User) error {
	ar.database.MutexUsers.Lock()
	defer ar.database.MutexUsers.Unlock()

	if _, exist := ar.database.UsersDB[user.ID]; exist {
		return repository.ErrorExistingUser
	}

	ar.database.UsersDB[user.ID] = user

	return nil
}

func (ar *AuthRepository) AddSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	ar.database.MutexSessions.Lock()
	defer ar.database.MutexSessions.Unlock()

	_, exist := ar.database.SessionsDB[sessionID]

	if exist {
		return repository.ErrorDetectingCollision
	}

	ar.database.SessionsDB[sessionID] = dbConnection.Session{
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return nil
}

func (ar *AuthRepository) GetUserIDBySession(ctx context.Context, sessionID string) (uuid.UUID, error) {
	ar.database.MutexSessions.Lock()
	defer ar.database.MutexSessions.Unlock()

	session, exist := ar.database.SessionsDB[sessionID]
	if !exist {
		return uuid.Nil, repository.ErrorNotExistingSession
	}

	if time.Now().After(session.ExpiresAt) {
		delete(ar.database.SessionsDB, sessionID)
		return uuid.Nil, repository.ErrorSeesionExpired
	}

	return session.UserID, nil
}

func (ar *AuthRepository) DeleteSession(ctx context.Context, sessionID string) error {
	ar.database.MutexSessions.Lock()
	defer ar.database.MutexSessions.Unlock()

	if _, exist := ar.database.SessionsDB[sessionID]; !exist {
		return repository.ErrorNotExistingSession
	}

	delete(ar.database.SessionsDB, sessionID)
	return nil
}

func (ar *AuthRepository) GetUser(ctx context.Context, email string) (models.User, error) {
	ar.database.MutexUsers.Lock()
	defer ar.database.MutexUsers.Unlock()

	for _, user := range ar.database.UsersDB {
		if user.Email == email {
			return user, nil
		}
	}

	return models.User{}, repository.ErrorNonexistentUser
}
