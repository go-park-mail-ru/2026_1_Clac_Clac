package repository

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/db"
	"github.com/google/uuid"
)

var (
	ErrorNotExistingSession = errors.New("session not found or expired")
	ErrorSeesionExpired     = errors.New("time life session expired")
)

type Repository struct {
	database *db.MapDatabases
}

func NewRepository(db *db.MapDatabases) *Repository {
	return &Repository{
		database: db,
	}
}

func (br *Repository) GetBoards(ctx context.Context, userID uuid.UUID) ([]models.Board, error) {
	br.database.MutexBoards.Lock()
	defer br.database.MutexBoards.Unlock()

	if user, exist := br.database.UsersDB[userID]; exist {

		newUser := models.User{
			Boards: []models.Board{},
		}

		for i := 0; i < len(user.Boards); i++ {
			newUser.Boards = append(newUser.Boards, models.Board(user.Boards[i]))
		}

		return newUser.Boards, nil
	}

	return nil, common.ErrorNonexistentUser
}

func (br *Repository) AddEmptyBoard(ctx context.Context, emptyBoard db.Board, userID uuid.UUID) error {
	br.database.MutexBoards.Lock()
	defer br.database.MutexBoards.Unlock()

	user, exist := br.database.UsersDB[userID]
	if !exist {
		return common.ErrorNonexistentUser
	}

	user.Boards = append(
		br.database.UsersDB[userID].Boards,
		emptyBoard,
	)

	br.database.UsersDB[userID] = user
	return nil
}
