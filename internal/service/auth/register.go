package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	repository "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/auth"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrorCreateHash = errors.New("failed to create hash")
)

type RegistrationService interface {
	Register(ctx context.Context, name, surname, password, email string) (models.User, error)
}

func CreateRegistrationService(repo repository.Database, hasher func(string) (string, error)) *RegistrationUserService {
	return &RegistrationUserService{
		repo:   repo,
		Hasher: hasher,
	}
}

func HashPassword(password string) (string, error) {
	sha256Hash := sha256.Sum256([]byte(password))
	hashString := hex.EncodeToString(sha256Hash[:])

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(hashString), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrorCreateHash, err)
	}

	return string(hashedPassword), nil
}

type RegistrationUserService struct {
	repo   repository.Database
	Hasher func(password string) (string, error)
}

func (r *RegistrationUserService) Register(ctx context.Context, name, surname, password, email string) (models.User, error) {
	hashedPassword, err := r.Hasher(password)
	if err != nil {
		return models.User{}, fmt.Errorf("HashPassword: %w", err)
	}

	user := models.User{
		ID:       uuid.New(),
		Name:     name,
		Surname:  surname,
		Password: hashedPassword,
		Email:    email,
		Boards:   make([]models.Board, 0),
	}

	err = r.repo.AddUser(ctx, user)
	if err != nil {
		return models.User{}, fmt.Errorf("repo.AddUser: %w", err)
	}

	return user, nil
}
