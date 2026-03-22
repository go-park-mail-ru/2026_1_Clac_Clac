package db

import (
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const timeBeforeRetry = time.Millisecond * 2

func TestNewPoolPostgresError(t *testing.T) {
	logger := zerolog.Nop()
	dbConfig := &config.DatabaseConnection{
		MinConnections:        1,
		MaxConnections:        5,
		MaxConnectionLifetime: time.Hour,
		MaxHealthCheckPeriod:  time.Minute,
	}

	tests := []struct {
		nameTest      string
		dsn           string
		expectedError error
		errorContains string
	}{
		{
			nameTest:      "Error invalid DSN format",
			dsn:           "invalid-dsn-string",
			expectedError: nil,
			errorContains: "cannot parse dsn",
		},
		{
			nameTest:      "Error connection timeout (exhaust retries)",
			dsn:           "postgres://user:pass@localhost:9999/testdb?sslmode=disable",
			expectedError: ErrorConnectPosgress,
			errorContains: "",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			pool, err := NewPoolPostgres(test.dsn, dbConfig, &logger, timeBeforeRetry)

			assert.Nil(t, pool, "pool should be nil on error")

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			}

			if test.errorContains != "" {
				assert.ErrorContains(t, err, test.errorContains)
			}
		})
	}
}

func TestRunMigrationsError(t *testing.T) {
	logger := zerolog.Nop()

	tests := []struct {
		nameTest      string
		dsn           string
		mockPath      string
		expectedError error
		errorContains string
	}{
		{
			nameTest:      "Error invalid database scheme",
			dsn:           "invalid-scheme://localhost",
			mockPath:      "file://migrations",
			errorContains: "cannot create migrate for database",
		},
		{
			nameTest:      "Error invalid migrations path",
			dsn:           "postgres://user:pass@localhost:5432/testdb?sslmode=disable",
			mockPath:      "file://invalid_path_that_does_not_exist",
			errorContains: "cannot create migrate for database",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			err := RunMigrations(test.dsn, &logger)

			require.Error(t, err)
			assert.ErrorContains(t, err, test.errorContains)
		})
	}
}
