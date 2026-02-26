package repository

import (
	"context"
	"errors"

	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
)

var ErrorExistingEmail = errors.New("user with this email alreday exists")

type Database interface {
	AddUser(ctx context.Context, user models.User) error
}

func NewMapDB() *MapDatabase {
	return &MapDatabase{
		database: make(map[string]models.User),
	}
}

type MapDatabase struct {
	database map[string]models.User
}

func (mp *MapDatabase) AddUser(ctx context.Context, user models.User) error {
	if _, exist := mp.database[user.Email]; exist {
		return ErrorExistingEmail
	}

	mp.database[user.Email] = user

	return nil
}
