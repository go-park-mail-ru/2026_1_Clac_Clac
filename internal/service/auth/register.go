package service

import (
	"context"
	"crypto/rand"
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
	Register(ctx context.Context, name, password, email string) (models.User, string, error)
}

func CreateRegistrationService(repo repository.Database, hasher func(string) (string, error), generatorSessionID func() (string, error)) *RegistrationUserService {
	return &RegistrationUserService{
		repo:        repo,
		Hasher:      hasher,
		GeneratorID: generatorSessionID,
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

func GenerateSessionID() (string, error) {
	buffer := make([]byte, 32)

	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("cannot generate sessinId: %w", err)
	}

	return hex.EncodeToString(buffer), nil
}

type RegistrationUserService struct {
	repo        repository.Database
	Hasher      func(password string) (string, error)
	GeneratorID func() (string, error)
}

func (r *RegistrationUserService) Register(ctx context.Context, name, password, email string) (models.User, string, error) {
	hashedPassword, err := r.Hasher(password)
	if err != nil {
		return models.User{}, "", fmt.Errorf("HashPassword: %w", err)
	}

	user := models.User{
		ID:          uuid.New(),
		DisplayName: name,
		Password:    hashedPassword,
		Email:       email,
	}

	err = r.repo.AddUser(ctx, user)
	if err != nil {
		return models.User{}, "", fmt.Errorf("repo.AddUser: %w", err)
	}

	sessionID, err := r.GeneratorID()
	if err != nil {
		return models.User{}, "", fmt.Errorf("GenerateSessionID: %w", err)
	}

	err = r.repo.AddSession(ctx, user.ID, sessionID)
	if err != nil {
		return models.User{}, "", fmt.Errorf("repo.AddSession: %w", err)
	}

	return user, sessionID, nil
}
