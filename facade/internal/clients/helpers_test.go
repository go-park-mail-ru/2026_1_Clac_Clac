package clients

import (
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestConvertGRPCError(t *testing.T) {
	plainErr := errors.New("plain error")

	tests := []struct {
		name     string
		input    error
		expected error
	}{
		{
			name:     "AlreadyExists",
			input:    status.Error(codes.AlreadyExists, "user already exists"),
			expected: common.ErrorExistingUser,
		},
		{
			name:     "NotFound with email identifier",
			input:    status.Error(codes.NotFound, "email not found"),
			expected: common.ErrorNonexistentEmail,
		},
		{
			name:     "NotFound with session identifier",
			input:    status.Error(codes.NotFound, "session expired"),
			expected: common.ErrorSessionNotFound,
		},
		{
			name:     "NotFound with reset token identifier",
			input:    status.Error(codes.NotFound, "reset token expired"),
			expected: common.ErrorResetTokenNotFound,
		},
		{
			name:     "NotFound default",
			input:    status.Error(codes.NotFound, "not found"),
			expected: common.ErrorNonexistentUser,
		},
		{
			name:     "InvalidArgument with wrong identifier",
			input:    status.Error(codes.InvalidArgument, "wrong credentials"),
			expected: common.ErrorWrongCredentials,
		},
		{
			name:     "InvalidArgument with null identifier",
			input:    status.Error(codes.InvalidArgument, "null field provided"),
			expected: common.ErrorNotNullValue,
		},
		{
			name:     "InvalidArgument default",
			input:    status.Error(codes.InvalidArgument, "bad input"),
			expected: common.ErrorInvalidInput,
		},
		{
			name:     "Unavailable",
			input:    status.Error(codes.Unavailable, "service unavailable"),
			expected: common.ErrorServiceUnavailable,
		},
		{
			name:     "Unknown gRPC code returns error as-is",
			input:    status.Error(codes.Internal, "internal error"),
			expected: status.Error(codes.Internal, "internal error"),
		},
		{
			name:     "Non-status error returns as-is",
			input:    plainErr,
			expected: plainErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertGRPCError(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
