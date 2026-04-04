package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateInfo(t *testing.T) {
	tests := []struct {
		nameTest      string
		info          string
		maxLen        int
		expectedError error
	}{
		{
			nameTest:      "Success validate name",
			info:          "bobr",
			maxLen:        maxLenNameUser,
			expectedError: nil,
		},
		{
			nameTest:      "Success validate description",
			info:          "bobrsjdjkaskjdn",
			maxLen:        maxLenDescriptionUser,
			expectedError: nil,
		},
		{
			nameTest:      "Error validate name",
			info:          "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			maxLen:        maxLenNameUser,
			expectedError: ErrorIncorrectLength,
		},
		{
			nameTest: "Error validate description",
			info: `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`,
			maxLen:        maxLenDescriptionUser,
			expectedError: ErrorIncorrectLength,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			err := ValidateInfo(test.info, test.maxLen)

			if test.expectedError != nil {
				assert.Equal(t, test.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
