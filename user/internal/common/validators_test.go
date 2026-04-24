package common

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAsciiSymbol(t *testing.T) {
	tests := []struct {
		nameTest string
		input    []string
		expected bool
	}{
		{
			nameTest: "All ASCII symbols",
			input:    []string{"hello", "world123", "test@mail.ru"},
			expected: true,
		},
		{
			nameTest: "Contains non-ASCII symbol",
			input:    []string{"hello", "мир"},
			expected: false,
		},
		{
			nameTest: "Empty strings",
			input:    []string{"", ""},
			expected: true,
		},
		{
			nameTest: "Single non-ASCII string",
			input:    []string{"пароль"},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			result := CheckAsciiSymbol(test.input...)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestValidateInfo(t *testing.T) {
	maxLenNameUser := 128
	maxLenDescriptionUser := 500

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
			expectedError: errors.New("must contain maximum 128 symbols"),
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
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`,
			maxLen:        maxLenDescriptionUser,
			expectedError: errors.New("must contain maximum 500 symbols"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			err := ValidateTextInfo(test.info, test.maxLen)

			if test.expectedError != nil {
				assert.Equal(t, test.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
