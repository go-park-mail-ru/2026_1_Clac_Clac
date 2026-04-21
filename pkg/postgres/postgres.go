package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func NewPoolPostgres(dsn string, conf *Config, logger *zerolog.Logger) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot parse dsn: %w", err)
	}

	poolConfig.MinConns = conf.MinConnections
	poolConfig.MaxConns = conf.MaxConnections
	poolConfig.MaxConnLifetime = conf.MaxConnectionLifetime
	poolConfig.HealthCheckPeriod = conf.MaxHealthCheckPeriod

	for i := 1; i <= conf.MaxRetries; i++ {
		contextWithTimeOut, cancel := context.WithTimeout(context.Background(), conf.TimeOut)

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

		if i < conf.MaxRetries {
			time.Sleep(conf.PingSleepTime)
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
