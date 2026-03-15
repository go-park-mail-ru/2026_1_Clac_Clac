package auth

import (
	"context"
	"time"

	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	db "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/db"
	"github.com/google/uuid"
)

type Repository struct {
	database *db.MapDatabases
}

func NewRepository(db *db.MapDatabases) *Repository {
	return &Repository{
		database: db,
	}
}

func (ar *Repository) AddUser(ctx context.Context, user models.User) error {
	ar.database.MutexUsers.Lock()
	defer ar.database.MutexUsers.Unlock()

	if _, exist := ar.database.UsersDB[user.ID]; exist {
		return common.ErrorExistingUser
	}

	dbBoards := []db.Board{}

	for i := 0; i < len(user.Boards); i++ {
		dbBoards = append(dbBoards, db.Board(user.Boards[i]))
	}

	ar.database.UsersDB[user.ID] = db.User{
		ID:           user.ID,
		DisplayName:  user.DisplayName,
		PasswordHash: user.PasswordHash,
		Email:        user.Email,
		Avatar:       user.Avatar,
		Boards:       dbBoards,
	}

	return nil
}

func (ar *Repository) AddSession(ctx context.Context, session db.Session) error {
	ar.database.MutexSessions.Lock()
	defer ar.database.MutexSessions.Unlock()

	_, exist := ar.database.SessionsDB[session.SessionID]
	if exist {
		return common.ErrorDetectingSessionCollision
	}

	ar.database.SessionsDB[session.SessionID] = session

	return nil
}

func (ar *Repository) GetUserIDBySession(ctx context.Context, sessionID string) (uuid.UUID, error) {
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

func (ar *Repository) DeleteSession(ctx context.Context, sessionID string) error {
	ar.database.MutexSessions.Lock()
	defer ar.database.MutexSessions.Unlock()

	if _, exist := ar.database.SessionsDB[sessionID]; !exist {
		return common.ErrorNotExistingSession
	}

	delete(ar.database.SessionsDB, sessionID)
	return nil
}

func (ar *Repository) GetUser(ctx context.Context, email string) (models.User, error) {
	ar.database.MutexUsers.RLock()
	defer ar.database.MutexUsers.RUnlock()

	for _, user := range ar.database.UsersDB {
		if user.Email == email {
			newUser := models.User{
				ID:           user.ID,
				DisplayName:  user.DisplayName,
				PasswordHash: user.PasswordHash,
				Email:        user.Email,
				Avatar:       user.Avatar,
				Boards:       []models.Board{},
			}

			for i := 0; i < len(user.Boards); i++ {
				newUser.Boards = append(newUser.Boards, models.Board(user.Boards[i]))
			}

			return newUser, nil
		}
	}

	return models.User{}, common.ErrorNonexistentUser
}

func (ar *Repository) GetResetToken(ctx context.Context, tokenID string) (db.ResetToken, error) {
	ar.database.MutexTokens.Lock()
	defer ar.database.MutexTokens.Unlock()

	token, exist := ar.database.ResetTokensDB[tokenID]
	if !exist {
		return db.ResetToken{}, common.ErrorNotExistingResetToken
	}

	if time.Now().After(token.ExpiresAt) {
		delete(ar.database.ResetTokensDB, tokenID)
		return db.ResetToken{}, common.ErrorResetTokenExpired
	}

	return token, nil
}

func (ar *Repository) DeleteResetToken(ctx context.Context, tokenID string) error {
	ar.database.MutexTokens.Lock()
	defer ar.database.MutexTokens.Unlock()

	_, exist := ar.database.ResetTokensDB[tokenID]
	if !exist {
		return common.ErrorNotExistingResetToken
	}

	delete(ar.database.ResetTokensDB, tokenID)

	return nil
}

func (ar *Repository) AddResetToken(ctx context.Context, token db.ResetToken) error {
	ar.database.MutexTokens.Lock()
	defer ar.database.MutexTokens.Unlock()

	_, exist := ar.database.ResetTokensDB[token.ResetTokenID]
	if exist {
		return common.ErrorDetectingTokenCollision
	}

	ar.database.ResetTokensDB[token.ResetTokenID] = token
	return nil
}

func (ar *Repository) UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error {
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
