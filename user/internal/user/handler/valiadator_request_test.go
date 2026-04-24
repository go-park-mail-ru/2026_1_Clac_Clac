package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	maxLenPassword = 128
	minLenPassword = 8
)

func TestValidatorRequestAuth(t *testing.T) {
	tests := []struct {
		nameTest      string
		email         string
		password      string
		expectedError error
	}{
		{
			nameTest:      "Success valid credentials",
			email:         "user@example.com",
			password:      "password123",
			expectedError: nil,
		},
		{
			nameTest:      "Error password too short",
			email:         "user@example.com",
			password:      "short",
			expectedError: ErrorLenPassword,
		},
		{
			nameTest:      "Error password too long",
			email:         "user@example.com",
			password:      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			expectedError: ErrorLenPassword,
		},
		{
			nameTest:      "Error invalid email format",
			email:         "not-an-email",
			password:      "password123",
			expectedError: ErrorIncorrectEmail,
		},
		{
			nameTest:      "Error non-ascii symbols in password",
			email:         "user@example.com",
			password:      "пароль123",
			expectedError: ErrorIncorrectSymbol,
		},
		{
			nameTest:      "Error non-ascii symbols in email",
			email:         "юзер@example.com",
			password:      "password123",
			expectedError: ErrorIncorrectSymbol,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			err := ValidatorRequestAuth(test.email, test.password, maxLenPassword, minLenPassword)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatorWithCheckPassword(t *testing.T) {
	tests := []struct {
		nameTest         string
		email            string
		password         string
		repeatedPassword string
		expectedError    error
	}{
		{
			nameTest:         "Success matching passwords",
			email:            "user@example.com",
			password:         "password123",
			repeatedPassword: "password123",
			expectedError:    nil,
		},
		{
			nameTest:         "Error passwords do not match",
			email:            "user@example.com",
			password:         "password123",
			repeatedPassword: "different456",
			expectedError:    ErrorDifferencePasswords,
		},
		{
			nameTest:         "Error invalid email with matching passwords",
			email:            "bad-email",
			password:         "password123",
			repeatedPassword: "password123",
			expectedError:    ErrorIncorrectEmail,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			err := ValidatorWithCheckPassword(test.email, test.password, test.repeatedPassword, maxLenPassword, minLenPassword)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatorRequestNewPassword(t *testing.T) {
	tests := []struct {
		nameTest         string
		password         string
		repeatedPassword string
		expectedError    error
	}{
		{
			nameTest:         "Success valid new password",
			password:         "newpassword1",
			repeatedPassword: "newpassword1",
			expectedError:    nil,
		},
		{
			nameTest:         "Error passwords do not match",
			password:         "newpassword1",
			repeatedPassword: "different",
			expectedError:    ErrorDifferencePasswords,
		},
		{
			nameTest:         "Error password too short",
			password:         "short",
			repeatedPassword: "short",
			expectedError:    ErrorLenPassword,
		},
		{
			nameTest:         "Error non-ascii symbols",
			password:         "пароль123",
			repeatedPassword: "пароль123",
			expectedError:    ErrorIncorrectSymbol,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			err := ValidatorRequestNewPassword(test.password, test.repeatedPassword, maxLenPassword, minLenPassword)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		nameTest string
		email    string
		expected bool
	}{
		{
			nameTest: "Valid email",
			email:    "user@example.com",
			expected: true,
		},
		{
			nameTest: "Valid email with subdomain",
			email:    "user@mail.example.com",
			expected: true,
		},
		{
			nameTest: "Invalid email without @",
			email:    "userexample.com",
			expected: false,
		},
		{
			nameTest: "Invalid empty string",
			email:    "",
			expected: false,
		},
		{
			nameTest: "Invalid email without domain",
			email:    "user@",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			result := ValidateEmail(test.email)
			assert.Equal(t, test.expected, result)
		})
	}
}
