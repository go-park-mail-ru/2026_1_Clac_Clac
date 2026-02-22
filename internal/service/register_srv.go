package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrorLenPassword      = errors.New("password must contain minimum 6 and maximum 18 characters")
	ErrorCreateHash       = errors.New("failed to create hash")
	ErrorCountAtSignEmail = errors.New("must use only one @ in email")
	ErrorIncorrectSymbol  = errors.New("invalid character: allowed only a-z, A-Z, 0-9, and /?!@")
)

type RegistrationService interface {
	Register(ctx context.Context, name, surname, password, email string) error
}

func CreateRegistrationService(repo repository.Database, hasher func(string) ([]byte, error)) *RegistrationUserService {
	return &RegistrationUserService{
		repo:   repo,
		Hasher: hasher,
	}
}

func HashPassword(password string) ([]byte, error) {
	sha256Hash := sha256.Sum256([]byte(password))
	hashString := hex.EncodeToString(sha256Hash[:])

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(hashString), bcrypt.DefaultCost)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %w", ErrorCreateHash, err)
	}

	return hashedPassword, nil
}

type RegistrationUserService struct {
	repo   repository.Database
	Hasher func(password string) ([]byte, error)
}

func CheckAsciiSymbol(strings ...string) bool {
	for _, str := range strings {
		for _, symbol := range str {
			if symbol > 127 {
				return false
			}
		}
	}

	return true
}

func (r *RegistrationUserService) Register(ctx context.Context, name, surname, password, email string) error {
	if isAsciiSymbol := CheckAsciiSymbol(name, surname, password, email); !isAsciiSymbol {
		return ErrorIncorrectSymbol
	}

	if len(password) < 6 {
		return ErrorLenPassword
	}

	countAtSign := strings.Count(email, "@")
	if countAtSign != 1 {
		return ErrorCountAtSignEmail
	}

	hashedPassword, err := r.Hasher(password)
	if err != nil {
		return fmt.Errorf("HashPassword: %w", err)
	}

	user := models.User{
		ID:       uuid.New(),
		Name:     name,
		Surname:  surname,
		Password: string(hashedPassword),
		Email:    email,
	}

	err = r.repo.AddUser(ctx, user)
	if err != nil {
		return fmt.Errorf("AddUser: %w", err)
	}

	return nil
}
