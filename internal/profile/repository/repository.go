package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	dbConnection "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/db"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/models"
	"github.com/google/uuid"
)

type Repository struct {
	database *dbConnection.MapDatabases
}

func NewProfileRepository(db *dbConnection.MapDatabases) *Repository {
	return &Repository{
		database: db,
	}
}

func (pr *Repository) GetProfile(ctx context.Context, userID uuid.UUID) (models.User, error) {
	pr.database.MutexUsers.Lock()
	defer pr.database.MutexUsers.Unlock()

	if user, exist := pr.database.UsersDB[userID]; exist {
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

	return models.User{}, common.ErrorNonexistentUser
}
