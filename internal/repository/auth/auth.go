package auth

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
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
		return common.ErrorExistingUser
	}

	ar.database.UsersDB[user.ID] = user

	return nil
}

func (ar *AuthRepository) AddSession(ctx context.Context, session dbConnection.Session) error {
	ar.database.MutexSessions.Lock()
	defer ar.database.MutexSessions.Unlock()

	_, exist := ar.database.SessionsDB[session.SessionID]
	if exist {
		return common.ErrorDetectingSessionCollision
	}

	ar.database.SessionsDB[session.SessionID] = session

	return nil
}

func (ar *AuthRepository) GetUserIDBySession(ctx context.Context, sessionID string) (uuid.UUID, error) {
	ar.database.MutexSessions.Lock()
	defer ar.database.MutexSessions.Unlock()

	session, exist := ar.database.SessionsDB[sessionID]
	if !exist {
		return uuid.Nil, common.ErrorNotExistingSession
	}

	if time.Now().After(session.ExpiresAt) {
		delete(ar.database.SessionsDB, sessionID)
		return uuid.Nil, common.ErrorSeesionExpired
	}

	return session.UserID, nil
}

func (ar *AuthRepository) DeleteSession(ctx context.Context, sessionID string) error {
	ar.database.MutexSessions.Lock()
	defer ar.database.MutexSessions.Unlock()

	if _, exist := ar.database.SessionsDB[sessionID]; !exist {
		return common.ErrorNotExistingSession
	}

	delete(ar.database.SessionsDB, sessionID)
	return nil
}

func (ar *AuthRepository) GetUser(ctx context.Context, email string) (models.User, error) {
	ar.database.MutexUsers.RLock()
	defer ar.database.MutexUsers.RUnlock()

	for _, user := range ar.database.UsersDB {
		if user.Email == email {
			return user, nil
		}
	}

	return models.User{}, common.ErrorNonexistentUser
}

func (ar *AuthRepository) GetResetToken(ctx context.Context, tokenID string) (dbConnection.ResetToken, error) {
	ar.database.MutexTokens.Lock()
	defer ar.database.MutexTokens.Unlock()

	token, exist := ar.database.ResetTokensDB[tokenID]
	if !exist {
		return dbConnection.ResetToken{}, common.ErrorNotExistingResetToken
	}

	if time.Now().After(token.ExpiresAt) {
		delete(ar.database.ResetTokensDB, tokenID)
		return dbConnection.ResetToken{}, common.ErrorResetTokenExpired
	}

	return token, nil
}

func (ar *AuthRepository) DeleteResetToken(ctx context.Context, tokenID string) error {
	ar.database.MutexTokens.Lock()
	defer ar.database.MutexTokens.Unlock()

	_, exist := ar.database.ResetTokensDB[tokenID]
	if !exist {
		return common.ErrorNotExistingResetToken
	}

	delete(ar.database.ResetTokensDB, tokenID)

	return nil
}

func (ar *AuthRepository) AddResetToken(ctx context.Context, token dbConnection.ResetToken) error {
	ar.database.MutexTokens.Lock()
	defer ar.database.MutexTokens.Unlock()

	_, exist := ar.database.ResetTokensDB[token.ResetTokenID]
	if exist {
		return common.ErrorDetectingTokenCollision
	}

	ar.database.ResetTokensDB[token.ResetTokenID] = token
	return nil
}

func (ar *AuthRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error {
	ar.database.MutexUsers.Lock()
	defer ar.database.MutexUsers.Unlock()

	user, exist := ar.database.UsersDB[userID]
	if !exist {
		return common.ErrorNonexistentUser
	}

	user.PasswordHash = newPasswordHash
	ar.database.UsersDB[userID] = user

	return nil
}
