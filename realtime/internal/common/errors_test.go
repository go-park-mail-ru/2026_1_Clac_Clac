package common_test

import (
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrTimeout", common.ErrTimeout},
		{"ErrBoardLinkMissing", common.ErrBoardLinkMissing},
		{"ErrInvalidBoardLink", common.ErrInvalidBoardLink},
		{"ErrParseLink", common.ErrParseLink},
		{"ErrUserNotAuthorized", common.ErrUserNotAuthorized},
		{"ErrCannotGetEvents", common.ErrCannotGetEvents},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc.err)
			assert.NotEmpty(t, tc.err.Error())
		})
	}
}
