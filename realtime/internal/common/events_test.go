package common_test

import (
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestBoardUpdateEvent(t *testing.T) {
	event := common.BoardUpdateEvent{
		BoardLink:  "board-link-123",
		EntityType: "card",
		EntityLink: "card-link-456",
		Action:     "create",
	}

	assert.Equal(t, "board-link-123", event.BoardLink)
	assert.Equal(t, "card", event.EntityType)
	assert.Equal(t, "card-link-456", event.EntityLink)
	assert.Equal(t, "create", event.Action)
}

func TestBoardUpdateEventZeroValue(t *testing.T) {
	event := common.BoardUpdateEvent{}
	assert.Empty(t, event.BoardLink)
	assert.Empty(t, event.EntityType)
	assert.Empty(t, event.EntityLink)
	assert.Empty(t, event.Action)
}
