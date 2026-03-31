package db

import (
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPoolPostgresError(t *testing.T) {
	const pingSleepTime = time.Millisecond * 2

	logger := zerolog.Nop()
	dbConfig := &config.DatabaseConnection{
		MinConnections:        1,
		MaxConnections:        5,
		MaxConnectionLifetime: time.Hour,
		MaxHealthCheckPeriod:  time.Minute,
		PingSleepTime:         pingSleepTime,
		TimeOut:               pingSleepTime,
		MaxRetries:            1,
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
			nameTest:      "Error connection timeout",
			dsn:           "postgres://user:pass@localhost:9999/testdb?sslmode=disable",
			expectedError: ErrorConnectPosgress,
			errorContains: "",
		},
		{
			nameTest:      "Error empty DSN string",
			dsn:           "",
			expectedError: ErrorConnectPosgress,
			errorContains: "",
		},
		{
			nameTest:      "Error unknown host in DSN",
			dsn:           "postgres://user:pass@unknown-host-that-does-not-exist:5432/testdb",
			expectedError: ErrorConnectPosgress,
			errorContains: "",
		},
		{
			nameTest:      "Error missing database name in DSN",
			dsn:           "postgres://user:pass@localhost:5432/",
			expectedError: ErrorConnectPosgress,
			errorContains: "",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			pool, err := NewPoolPostgres(test.dsn, dbConfig, &logger)

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
		{
			nameTest:      "Error empty DSN string",
			dsn:           "",
			expectedError: nil,
			errorContains: "cannot create migrate for database",
		},
		{
			nameTest:      "Error malformed postgres URL",
			dsn:           "postgres://user@:pass@localhost/db",
			mockPath:      "file://migrations",
			errorContains: "cannot create migrate for database",
		},
		{
			nameTest:      "Error unreachable host with fast timeout",
			dsn:           "postgres://user:pass@10.255.255.1:5432/db?connect_timeout=1",
			mockPath:      "file://migrations",
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
