package common

import (
	"errors"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

var (
	ErrorExistingUser     = errors.New("user with this email alreday exists")
	ErrorNotNullValue     = errors.New("put null value in not null field")
	ErrorNonexistentUser  = errors.New("user with this ID not exist")
	ErrorNonexistentEmail = errors.New("user with this email not exist")

	ErrorDetectingSessionCollision = errors.New("session collision detected")

	ErrorNotExistingSession = errors.New("session not found or expired")
	ErrorSeesionExpired     = errors.New("time life session expired")

	ErrorDecodeRequest = errors.New("decoding request is incorrect")

	ErrorNotExistingResetToken   = errors.New("reset token not found or expire")
	ErrorResetTokenExpired       = errors.New("time life reset token expired")
	ErrorDetectingTokenCollision = errors.New("reset token collision detected")

	ErrorExistingBoard = errors.New("board with this ID alreday exists")

	ErrorNotExistingSection = errors.New("section not found")
	ErrorDeleteBacklog      = errors.New("can not delete backlog section")
	ErrorUpdateBacklog      = errors.New("can not update backlog section")
	ErrorNotFindAllLinks    = errors.New("not all sections have found")

	ErrorNotExistingCard          = errors.New("task not found")
	ErrorCardAlreadyExist         = errors.New("card with this link is already exist")
	ErrorInvalidReferenceCardData = errors.New("invalid references for card")
	ErrorInvalidCardData          = errors.New("incorrect card data")

	ErrorMissingRequiredField = errors.New("required field is missing")

	ErrorSkipMandatorySection        = errors.New("can not move through mandatory sections")
	ErrorInvalidSectionData          = errors.New("incorrect section data")
	ErrorSectionAlreadyExist         = errors.New("section is already exist")
	ErrorInvalidReferenceSectionData = errors.New("invalid references for section")

	ErrorInvalidProfileData = errors.New("incorrect profile data")
	ErrorIncorrectSymbol    = errors.New("allowed only a-z, A-Z, 0-9, and /?!@")
	ErrorIncorrectColor     = errors.New("color is incorrect, can be white, grey, red, orange, blue, green, purple, pink")
)
