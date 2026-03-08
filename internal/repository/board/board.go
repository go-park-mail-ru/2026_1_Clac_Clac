package board

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	dbConnection "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/db_connection"
	"github.com/google/uuid"
)

var (
	ErrorNotExistingSession = errors.New("session not found or expired")
	ErrorSeesionExpired     = errors.New("time life session expired")
)

type BoardRepositpry struct {
	database *dbConnection.MapDatabases
}

func NewBoardRepository(db *dbConnection.MapDatabases) *BoardRepositpry {
	return &BoardRepositpry{
		database: db,
	}
}

func (br *BoardRepositpry) GetBoards(ctx context.Context, userID uuid.UUID) []models.Board {
	br.database.MutexBoards.Lock()
	defer br.database.MutexBoards.Unlock()

	for _, user := range br.database.UsersDB {
		if user.ID == userID {
			return user.Boards
		}
	}

	return nil
}
