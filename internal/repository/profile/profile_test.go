package profile

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	dbConnection "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/db_connection"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetUser(t *testing.T) {
	tests := []struct {
		nameTest     string
		userID       uuid.UUID
		mapUsers     map[uuid.UUID]models.User
		expectedUser models.User
	}{
		{
			nameTest: "Success get user",
			userID:   common.FixedUserUuiD,
			mapUsers: map[uuid.UUID]models.User{common.FixedUserUuiD: {ID: common.FixedUserUuiD}},
			expectedUser: models.User{
				ID: common.FixedUserUuiD,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.MapDatabases{UsersDB: test.mapUsers}
			repoUsers := NewProfileRepository(&conectionDb)

			ctx := context.Background()

			user, _ := repoUsers.GetProfile(ctx, test.userID)

			assert.Equal(t, test.expectedUser, user)
		})
	}
}
