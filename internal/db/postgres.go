package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var ErrorConnectPosgress = errors.New("cannot connected to Postgres")

const (
	migrationsPath = "file://internal/db/migrations"
)

func NewPoolPostgres(dsn string, dbConnection *config.DatabaseConnection, logger *zerolog.Logger) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot parse dsn: %w", err)
	}

	poolConfig.MinConns = dbConnection.MinConnections
	poolConfig.MaxConns = dbConnection.MaxConnections
	poolConfig.MaxConnLifetime = dbConnection.MaxConnectionLifetime
	poolConfig.HealthCheckPeriod = dbConnection.MaxHealthCheckPeriod

	for i := 1; i <= dbConnection.MaxRetries; i++ {
		contextWithTimeOut, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		if pool, err := pgxpool.NewWithConfig(contextWithTimeOut, poolConfig); err == nil {
			pingErr := pool.Ping(contextWithTimeOut)
			if pingErr == nil {
				cancel()
				logger.Info().Msgf("Successfully connected to Postgres (Attempt %d)", i)
				return pool, nil
			}

			pool.Close()
		}

		logger.Warn().Msgf("Postgres not ready yet, retrying")

		if i < dbConnection.MaxRetries {
			time.Sleep(dbConnection.PingSleepTime)
		}
		cancel()
	}

	return nil, ErrorConnectPosgress
}

func RunMigrations(dsn string, logger *zerolog.Logger) (err error) {
	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("cannot create migrate for database: %w", err)
	}

	defer func() {
		errSource, errDB := m.Close()
		if errSource != nil {
			errSource = fmt.Errorf("cannot close migrations source: %w", errSource)
			err = errors.Join(err, errSource)
		}

		if errDB != nil {
			errDb := fmt.Errorf("cannot close migrations db: %w", errDB)
			err = errors.Join(err, errDb)
		}
	}()

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info().Msg("No change for database")
			return nil
		}

		return fmt.Errorf("cannot up migrations: %w", err)
	}

	logger.Info().Msg("Migrations applied successfully")
	return nil
}
