package common_test

import (
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestBoardUpdateEvent(t *testing.T) {
	event := common.BoardUpdateEvent{
		BoardLink: "board-link-123",
		UserLink:  "user-link-456",
		Data:      map[string]string{"key": "value"},
	}

	assert.Equal(t, "board-link-123", event.BoardLink)
	assert.Equal(t, "user-link-456", event.UserLink)
	assert.Equal(t, map[string]string{"key": "value"}, event.Data)
}

func TestBoardUpdateEventZeroValue(t *testing.T) {
	event := common.BoardUpdateEvent{}
	assert.Empty(t, event.BoardLink)
	assert.Empty(t, event.UserLink)
	assert.Nil(t, event.Data)
}
