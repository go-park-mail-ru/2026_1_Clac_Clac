package repository

import (
	"context"
	"testing"

	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestAddUser(t *testing.T) {
	tests := []struct {
		nameTest      string
		emails        []string
		expectedError error
	}{
		{
			nameTest:      "Success registration",
			emails:        []string{"bobr@mail.ru"},
			expectedError: nil,
		},
		{
			nameTest:      "Email is already existing",
			emails:        []string{"bobr@mail.ru", "bobr@mail.ru"},
			expectedError: ErrorExistingEmail,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			repoUsers := NewMapDB()

			var err error
			ctx := context.Background()

			for _, email := range test.emails {
				err = repoUsers.AddUser(ctx, models.User{Email: email})
			}

			assert.Equal(t, test.expectedError, err)
		})
	}
}
