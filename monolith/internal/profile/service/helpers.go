package service

import "github.com/google/uuid"

func GenerateAvatarKey() (string, error) {
	return uuid.New().String(), nil
}
