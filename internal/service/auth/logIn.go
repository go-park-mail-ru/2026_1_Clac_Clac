package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	repository "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/auth"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrorWrongPassword = errors.New("write wrong password")
)

type LogInService interface {
	Login(ctx context.Context, email, password string) (models.User, string, error)
}

type LogInUserService struct {
	repo        repository.Database
	checker     func(string, string) error
	generatorID func() (string, error)
}

func NewLogInService(repo repository.Database, checker func(string, string) error, generatorSessionID func() (string, error)) *LogInUserService {
	return &LogInUserService{
		repo:        repo,
		checker:     checker,
		generatorID: generatorSessionID,
	}
}

func CheckPassword(inputPassword, hashPassword string) error {
	sha256Hash := sha256.Sum256([]byte(inputPassword))
	inputHashString := hex.EncodeToString(sha256Hash[:])

	if err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(inputHashString)); err != nil {
		return ErrorWrongPassword
	}

	return nil
}

func (l *LogInUserService) Login(ctx context.Context, email, password string) (models.User, string, error) {
	user, err := l.repo.GetUser(ctx, email)
	if err != nil {
		return models.User{}, "", fmt.Errorf("repo.GetUser: %w", err)
	}

	err = l.checker(password, user.PasswordHash)
	if err != nil {
		return models.User{}, "", fmt.Errorf("repo.CheckPassword: %w", err)
	}

	sessionID, err := l.generatorID()
	if err != nil {
		return models.User{}, "", fmt.Errorf("GenerateID: %w", err)
	}

	err = l.repo.AddSession(ctx, user.ID, sessionID)
	if err != nil {
		return models.User{}, "", fmt.Errorf("repo.AddSession: %w", err)
	}

	return user, sessionID, nil
}
