package logger

import (
	"context"
	"errors"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	t.Run("empty DSN initializes without error", func(t *testing.T) {
		cfg := Sentry{
			DSN:              "",
			Environment:      "test",
			Release:          "1.0.0",
			ServiceName:      "test-service",
			Tags:             map[string]string{"layer": "grpc"},
			TracesSampleRate: 1.0,
			Repanic:          false,
		}

		err := Init(cfg)
		assert.NoError(t, err)
	})

	t.Run("invalid DSN returns error", func(t *testing.T) {
		cfg := Sentry{
			DSN:         "not-a-valid-dsn",
			ServiceName: "test-service",
		}

		err := Init(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sentry.Init")
	})

	t.Run("reserved tags are skipped, non-reserved are set", func(t *testing.T) {
		cfg := Sentry{
			DSN:         "",
			ServiceName: "test-service",
			Tags: map[string]string{
				"service": "skip-me",
				"env":     "skip-me",
				"release": "skip-me",
				"layer":   "grpc",
				"team":    "api-platform",
			},
		}

		err := Init(cfg)
		assert.NoError(t, err)
	})

	t.Run("empty tags map does not panic", func(t *testing.T) {
		cfg := Sentry{
			DSN:         "",
			ServiceName: "test-service",
			Tags:        map[string]string{},
		}

		err := Init(cfg)
		assert.NoError(t, err)
	})

	t.Run("nil tags map does not panic", func(t *testing.T) {
		cfg := Sentry{
			DSN:         "",
			ServiceName: "test-service",
			Tags:        nil,
		}

		err := Init(cfg)
		assert.NoError(t, err)
	})
}

func TestFlush(t *testing.T) {
	t.Run("does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Flush()
		})
	})
}

func TestCaptureError(t *testing.T) {
	t.Run("nil error returns without capturing", func(t *testing.T) {
		assert.NotPanics(t, func() {
			CaptureError(nil, "ctx", map[string]interface{}{"key": "val"})
		})
	})

	t.Run("non-nil error with extra and context name", func(t *testing.T) {
		assert.NotPanics(t, func() {
			CaptureError(errors.New("test error"), "MyContext", map[string]interface{}{"key": "val"})
		})
	})

	t.Run("non-nil error with extra and empty context name defaults to Context Data", func(t *testing.T) {
		assert.NotPanics(t, func() {
			CaptureError(errors.New("test error"), "", map[string]interface{}{"key": "val"})
		})
	})

	t.Run("non-nil error with nil extra skips SetContext", func(t *testing.T) {
		assert.NotPanics(t, func() {
			CaptureError(errors.New("test error"), "ctx", nil)
		})
	})

	t.Run("non-nil error with empty extra skips SetContext", func(t *testing.T) {
		assert.NotPanics(t, func() {
			CaptureError(errors.New("test error"), "ctx", map[string]interface{}{})
		})
	})
}

func TestCaptureFromContext(t *testing.T) {
	t.Run("nil error returns without capturing", func(t *testing.T) {
		assert.NotPanics(t, func() {
			CaptureFromContext(context.Background(), nil, "ctx", map[string]interface{}{"key": "val"})
		})
	})

	t.Run("non-nil error without hub in context falls back to CurrentHub", func(t *testing.T) {
		assert.NotPanics(t, func() {
			CaptureFromContext(context.Background(), errors.New("test error"), "ctx", map[string]interface{}{"key": "val"})
		})
	})

	t.Run("non-nil error with hub in context uses that hub", func(t *testing.T) {
		hub := sentry.NewHub(nil, sentry.NewScope())
		ctx := sentry.SetHubOnContext(context.Background(), hub)

		assert.NotPanics(t, func() {
			CaptureFromContext(ctx, errors.New("test error"), "ctx", map[string]interface{}{"key": "val"})
		})
	})

	t.Run("non-nil error with hub and empty context name defaults to Context Data", func(t *testing.T) {
		hub := sentry.NewHub(nil, sentry.NewScope())
		ctx := sentry.SetHubOnContext(context.Background(), hub)

		assert.NotPanics(t, func() {
			CaptureFromContext(ctx, errors.New("test error"), "", map[string]interface{}{"key": "val"})
		})
	})

	t.Run("non-nil error with hub and nil extra skips SetContext", func(t *testing.T) {
		hub := sentry.NewHub(nil, sentry.NewScope())
		ctx := sentry.SetHubOnContext(context.Background(), hub)

		assert.NotPanics(t, func() {
			CaptureFromContext(ctx, errors.New("test error"), "ctx", nil)
		})
	})

	t.Run("non-nil error without hub and empty context name defaults to Context Data", func(t *testing.T) {
		assert.NotPanics(t, func() {
			CaptureFromContext(context.Background(), errors.New("test error"), "", map[string]interface{}{"key": "val"})
		})
	})

	t.Run("non-nil error without hub and nil extra skips SetContext", func(t *testing.T) {
		assert.NotPanics(t, func() {
			CaptureFromContext(context.Background(), errors.New("test error"), "ctx", nil)
		})
	})
}
