package config

import (
	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
	"github.com/spf13/viper"
)

const (
	defaultDSN              = ""
	defaultEnvironment      = "production"
	defaultRelease          = "RK3"
	defaultServiceName      = "facade"
	defaultTracesSampleRate = 0.1
	defaultRepanic          = true
)

var (
	defaultTags = map[string]string{
		"layer":    "edge",
		"team":     "api-platform",
		"protocol": "http_rest",
	}
)

func DefaultSentryConfig() sentryLogger.Sentry {
	return sentryLogger.Sentry{
		Environment:      defaultEnvironment,
		Release:          defaultRelease,
		ServiceName:      defaultServiceName,
		Tags:             defaultTags,
		TracesSampleRate: defaultTracesSampleRate,
		Repanic:          defaultRepanic,
	}
}

func SetupEnvSentryConfig(v *viper.Viper) {
	v.SetDefault("sentry.dsn", defaultDSN)
	v.RegisterAlias("sentry.dsn", "sentry_dsn")
}
