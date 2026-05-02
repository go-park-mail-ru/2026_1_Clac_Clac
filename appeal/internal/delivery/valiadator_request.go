package delivery

import (
	"errors"
	"net/mail"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/common"
)

var (
	ErrorLenPassword         = errors.New("password must contain minimum 8 and maximum 128 symbols")
	ErrorIncorrectEmail      = errors.New("invalid email format")
	ErrorIncorrectSymbol     = errors.New("allowed only a-z, A-Z, 0-9, and /?!@")
	ErrorDifferencePasswords = errors.New("passwords don't match")
)

const (
	maxLenName = 128
)

func ValidatorRequestAppeal(email string, name string) error {
	correctSymbols := common.CheckAsciiSymbol(email)
	if !correctSymbols {
		return ErrorIncorrectSymbol
	}

	if len(name) > maxLenName {
		return ErrorLenPassword
	}

	correctEmail := ValidateEmail(email)
	if !correctEmail {
		return ErrorIncorrectEmail
	}

	return nil
}

func ValidateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
