package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSessionID(t *testing.T) {
	id1, err := GenerateSessionID()
	assert.NoError(t, err, "expected no error while generating session ID")
	assert.Equal(t, 64, len(id1), "hex encoded array should be 64 characters long")

	id2, err := GenerateSessionID()
	assert.NoError(t, err)
	assert.NotEqual(t, id1, id2, "generated sessionID should be unique")
}
