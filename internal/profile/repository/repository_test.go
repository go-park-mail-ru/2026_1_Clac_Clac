package repository

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	db "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetUser(t *testing.T) {
	tests := []struct {
		nameTest     string
		userID       uuid.UUID
		mapUsers     map[uuid.UUID]db.User
		expectedUser db.User
	}{
		{
			nameTest: "Success get user",
			userID:   common.FixedUserUuiD,
			mapUsers: map[uuid.UUID]db.User{common.FixedUserUuiD: {ID: common.FixedUserUuiD}},
			expectedUser: db.User{
				ID:     common.FixedUserUuiD,
				Boards: []db.Board{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := db.MapDatabases{UsersDB: test.mapUsers}
			repoUsers := NewProfileRepository(&conectionDb)

			ctx := context.Background()

			user, _ := repoUsers.GetProfile(ctx, test.userID)

			newUser := db.User{
				ID:           user.ID,
				DisplayName:  user.DisplayName,
				PasswordHash: user.PasswordHash,
				Email:        user.Email,
				Avatar:       user.Avatar,
				Boards:       []db.Board{},
			}

			for i := 0; i < len(user.Boards); i++ {
				newUser.Boards = append(newUser.Boards, db.Board(user.Boards[i]))
			}

			assert.Equal(t, test.expectedUser, newUser)
		})
	}
}
