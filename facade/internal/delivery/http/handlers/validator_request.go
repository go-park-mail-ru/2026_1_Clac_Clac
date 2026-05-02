package handlers

import (
	"errors"
	"net/mail"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
)

var (
	ErrorLenPassword         = errors.New("password must contain minimum 8 and maximum 128 symbols")
	ErrorIncorrectEmail      = errors.New("invalid email format")
	ErrorIncorrectSymbol     = errors.New("allowed only a-z, A-Z, 0-9, and /?!@")
	ErrorDifferencePasswords = errors.New("passwords don't match")
)

func ValidatorWithCheckPassword(email, password, repeatedPassword string, maxLenPassword, minLenPassword int) error {
	if password != repeatedPassword {
		return ErrorDifferencePasswords
	}

	return ValidatorRequestAuth(email, password, maxLenPassword, minLenPassword)
}

func ValidatorRequestAuth(email, password string, maxLenPassword, minLenPassword int) error {
	correctSymbols := common.CheckAsciiSymbol(email, password)
	if !correctSymbols {
		return ErrorIncorrectSymbol
	}

	if len(password) < minLenPassword || len(password) > maxLenPassword {
		return ErrorLenPassword
	}

	correctEmail := ValidateEmail(email)
	if !correctEmail {
		return ErrorIncorrectEmail
	}

	return nil
}

func ValidatorRequestNewPassword(password, repeatedPassword string, maxLenPassword, minLenPassword int) error {
	if password != repeatedPassword {
		return ErrorDifferencePasswords
	}

	correctSymbols := common.CheckAsciiSymbol(password)
	if !correctSymbols {
		return ErrorIncorrectSymbol
	}

	if len(password) < minLenPassword || len(password) > maxLenPassword {
		return ErrorLenPassword
	}

	return nil
}

func ValidateEmail(email string) bool {
	field, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	return field.Address == email
}
