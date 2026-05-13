package clients

import (
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	identifierEmailError      = "email"
	identifierSessionError    = "session"
	identifierResetTokenError = "reset token"
	identifierWrongError      = "wrong"
	identifierNullFieldError  = "null"

	identifierCardNotFound       = "card not found"
	identifierSectionNotFound    = "section not found"
	identifierBoardNotFound      = "board not found"
	identifierCommentNotFound    = "comment not found"
	identifierSubtaskNotFound    = "sub task not found"
	identifierAttachmentNotFound = "attachment not found"

	identifierTaskLimitError       = "task limit"
	identifierLostMandatorySection = "mandatory section"
	identifierIncorrectTypeFile    = "invalid content type"

	identifierAppealNotFound  = "appeal not found"
	identifierInvalidCategory = "invalid category"
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
		case strings.Contains(msg, identifierIncorrectTypeFile):
			return common.ErrorInvalidContentType
		default:
			return common.ErrorInvalidInput
		}
	case codes.Unavailable:
		return common.ErrorVKOAuthUnavailable
	default:
		return err
	}
}

func convertCardGRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	msg := st.Message()
	switch st.Code() {
	case codes.NotFound:
		switch {
		case strings.Contains(msg, identifierCardNotFound):
			return common.ErrorCardNotFound
		case strings.Contains(msg, identifierSectionNotFound):
			return common.ErrorSectionNotFound
		case strings.Contains(msg, identifierCommentNotFound):
			return common.ErrorCommentNotFound
		case strings.Contains(msg, identifierSubtaskNotFound):
			return common.ErrorSubtaskNotFound
		case strings.Contains(msg, identifierAttachmentNotFound):
			return common.ErrorAttachmentNotFound
		default:
			return common.ErrorNonexistentUser
		}
	case codes.PermissionDenied:
		return common.ErrorPermissionDenied
	case codes.AlreadyExists:
		return common.ErrorCardAlreadyExists
	case codes.InvalidArgument:
		switch {
		case strings.Contains(msg, identifierTaskLimitError):
			return common.ErrorTaskLimitReached
		case strings.Contains(msg, identifierLostMandatorySection):
			return common.ErrCannotSkipMandatorySection
		}

		return common.ErrorInvalidInput
	default:
		return err
	}
}

func convertBoardGRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	msg := st.Message()
	switch st.Code() {
	case codes.NotFound:
		switch {
		case strings.Contains(msg, identifierBoardNotFound):
			return common.ErrorBoardNotFound
		case strings.Contains(msg, identifierSectionNotFound):
			return common.ErrorSectionNotFound
		default:
			return common.ErrorNonexistentUser
		}
	case codes.InvalidArgument:
		switch {
		case strings.Contains(msg, identifierIncorrectTypeFile):
			return common.ErrorInvalidContentType
		default:
			return common.ErrorInvalidInput
		}
	case codes.PermissionDenied:
		return common.ErrorBoardPermissionDenied
	default:
		return err
	}
}

func convertSectionGRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	msg := st.Message()
	switch st.Code() {
	case codes.NotFound:
		switch {
		case strings.Contains(msg, identifierSectionNotFound):
			return common.ErrorSectionNotFound
		case strings.Contains(msg, identifierBoardNotFound):
			return common.ErrorBoardNotFound
		default:
			return common.ErrorNonexistentUser
		}
	case codes.PermissionDenied:
		return common.ErrorSectionPermissionDenied
	default:
		return err
	}
}

func convertAppealGRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	msg := st.Message()
	switch st.Code() {
	case codes.AlreadyExists:
		return common.ErrorExistingUser
	case codes.NotFound:
		switch {
		case strings.Contains(msg, identifierAppealNotFound):
			return common.ErrorAppealNotFound
		default:
			return common.ErrorNonexistentUser
		}
	case codes.PermissionDenied:
		return common.ErrorPermissionDenied
	case codes.InvalidArgument:
		switch {
		case strings.Contains(msg, identifierInvalidCategory):
			return common.ErrInvalidCategory
		case strings.Contains(msg, identifierNullFieldError):
			return common.ErrorNotNullValue
		default:
			return common.ErrorInvalidInput
		}
	default:
		return err
	}
}

func convertTimeToTimestamppb(t *time.Time) *timestamppb.Timestamp {
	if t != nil {
		return timestamppb.New(*t)
	}
	return nil
}

func convertTimestamppbToTime(time *timestamppb.Timestamp) *time.Time {
	if time != nil {
		convertedTime := time.AsTime()

		return &convertedTime
	}

	return nil
}
