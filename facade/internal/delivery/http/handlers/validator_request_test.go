package handlers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	maxLen = 128
	minLen = 8
)

func TestValidatorRequestAuth(t *testing.T) {
	tests := []struct {
		Name          string
		Email         string
		Password      string
		ExpectedError error
	}{
		{
			Name:          "Success validation",
			Email:         "bobr@mail.ru",
			Password:      "secure_password",
			ExpectedError: nil,
		},
		{
			Name:          "Invalid email format",
			Email:         "invalid-email-format",
			Password:      "secure_password",
			ExpectedError: ErrorIncorrectEmail,
		},
		{
			Name:          "Password too short",
			Email:         "bobr@mail.ru",
			Password:      "1234567",
			ExpectedError: ErrorLenPassword,
		},
		{
			Name:          "Password too long",
			Email:         "bobr@mail.ru",
			Password:      strings.Repeat("a", 129),
			ExpectedError: ErrorLenPassword,
		},
		{
			Name:          "Not ASCII symbol in password",
			Email:         "valid.user@example.com",
			Password:      "пароль123",
			ExpectedError: ErrorIncorrectSymbol,
		},
		{
			Name:          "Not ASCII symbol in email",
			Email:         "почта@mail.ru",
			Password:      "passsssword",
			ExpectedError: ErrorIncorrectSymbol,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := ValidatorRequestAuth(test.Email, test.Password, maxLen, minLen)

			assert.Equal(t, test.ExpectedError, err, "incorrect error")
		})
	}
}

func TestValidatorRequestNewPassword(t *testing.T) {
	tests := []struct {
		Name             string
		Password         string
		RepeatedPassword string
		ExpectedError    error
	}{
		{
			Name:             "Success validation",
			Password:         "passsssword",
			RepeatedPassword: "passsssword",
			ExpectedError:    nil,
		},
		{
			Name:             "Passwords do not match",
			Password:         "passsssword",
			RepeatedPassword: "passssssssword",
			ExpectedError:    ErrorDifferencePasswords,
		},
		{
			Name:             "Password too short",
			Password:         "short",
			RepeatedPassword: "short",
			ExpectedError:    ErrorLenPassword,
		},
		{
			Name:             "Not ASCII symbol in password",
			Password:         "бобр_password!",
			RepeatedPassword: "бобр_password!",
			ExpectedError:    ErrorIncorrectSymbol,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := ValidatorRequestNewPassword(test.Password, test.RepeatedPassword, maxLen, minLen)

			assert.Equal(t, test.ExpectedError, err, "incorrect error")
		})
	}
}

func TestValidatorWithCheckPassword(t *testing.T) {
	tests := []struct {
		name             string
		email            string
		password         string
		repeatedPassword string
		expectedError    error
	}{
		{
			name:             "Passwords do not match",
			email:            "bobr@mail.ru",
			password:         "pass1234",
			repeatedPassword: "pass5678",
			expectedError:    ErrorDifferencePasswords,
		},
		{
			name:             "Success delegates to ValidatorRequestAuth",
			email:            "bobr@mail.ru",
			password:         "pass1234",
			repeatedPassword: "pass1234",
			expectedError:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatorWithCheckPassword(tc.email, tc.password, tc.repeatedPassword, maxLen, minLen)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}
