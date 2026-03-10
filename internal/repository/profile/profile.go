package profile

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	dbConnection "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/db_connection"
	"github.com/google/uuid"
)

type ProfileRepository struct {
	database *dbConnection.MapDatabases
}

func NewProfileRepository(db *dbConnection.MapDatabases) *ProfileRepository {
	return &ProfileRepository{
		database: db,
	}
}

func (pr *ProfileRepository) GetProfile(ctx context.Context, userID uuid.UUID) (models.User, error) {
	pr.database.MutexUsers.Lock()
	defer pr.database.MutexUsers.Unlock()

	if user, exist := pr.database.UsersDB[userID]; exist {
		return user, nil
	}

	return models.User{}, common.ErrorNonexistentUser
}
