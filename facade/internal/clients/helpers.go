package clients

import (
	"strings"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	identifierEmailError      = "email"
	identifierSessionError    = "session"
	identifierResetTokenError = "reset token"
	identifierWrongError      = "wrong"
	identifierNullFieldError  = "null"
)

func convertGRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch st.Code() {
	case codes.AlreadyExists:
		return common.ErrorExistingUser
	case codes.NotFound:
		msg := st.Message()
		switch {
		case strings.Contains(msg, identifierEmailError):
			return common.ErrorNonexistentEmail
		case strings.Contains(msg, identifierSessionError):
			return common.ErrorSessionNotFound
		case strings.Contains(msg, identifierResetTokenError):
			return common.ErrorResetTokenNotFound
		default:
			return common.ErrorNonexistentUser
		}
	case codes.InvalidArgument:
		msg := st.Message()
		switch {
		case strings.Contains(msg, identifierWrongError):
			return common.ErrorWrongCredentials
		case strings.Contains(msg, identifierNullFieldError):
			return common.ErrorNotNullValue
		default:
			return common.ErrorInvalidInput
		}
	case codes.Unavailable:
		return common.ErrorVKOAuthUnavailable
	default:
		return err
	}
}
