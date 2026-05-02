package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
)

func Init(cfg Sentry) error {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.DSN,
		Environment: cfg.Environment,
		Release:     cfg.Release,
		ServerName:  cfg.ServiceName,
		// TracesSampleRate: 0.2,

		// BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
		// 	return event
		// },
	})

	if err != nil {
		return fmt.Errorf("sentry.Init: %W", err)
	}

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("service", cfg.ServiceName)

		for key, value := range cfg.Tags {
			if key != "service" && key != "env" && key != "release" {
				scope.SetTag(key, value)
			}
		}
	})

	return nil
}

func Flush() {
	sentry.Flush(2 * time.Second)
}

func CaptureError(err error, contextName string, extra map[string]interface{}) {
	if err == nil {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		if len(extra) > 0 {
			if contextName == "" {
				contextName = "Context Data"
			}
			scope.SetContext(contextName, extra)
		}
		sentry.CaptureException(err)
	})
}

func CaptureFromContext(ctx context.Context, err error, contextName string, extra map[string]interface{}) {
	if err == nil {
		return
	}

	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	hub.WithScope(func(scope *sentry.Scope) {
		if len(extra) > 0 {
			if contextName == "" {
				if contextName == "" {
					contextName = "Context Data"
				}
				scope.SetContext(contextName, extra)
			}
		}
	})
}
