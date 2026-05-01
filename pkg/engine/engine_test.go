package engine_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/engine"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestStartListenError(t *testing.T) {
	cfg := engine.Config{
		Addr: "invalid-address-that-cannot-listen",
	}
	buf := &bytes.Buffer{}
	logger := zerolog.New(buf)
	router := mux.NewRouter()
	e := engine.New(&cfg, &logger, router)

	err := e.Start(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to listen")
}

func TestGracefulShutdown(t *testing.T) {
	const timeout = 16 * time.Second
	cfg := engine.Config{
		Addr:                    ":0",
		WriteTimeout:            15,
		ReadTimeout:             15,
		IdleTimeout:             60,
		GracefulShutdownTimeout: 15,
	}

	buf := &bytes.Buffer{}
	logger := zerolog.New(buf)
	router := mux.NewRouter()

	e := engine.New(&cfg, &logger, router)

	serverReady := make(chan struct{})
	e.OnListen = func(_ string) {
		close(serverReady)
	}

	serverStopped := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		e.Start(ctx)
		close(serverStopped)
	}()

	<-serverReady
	// Сервер запущен, можно делать shutdown

	cancel()

	select {
	case <-serverStopped:
	case <-time.After(timeout):
		t.Fatal("server did not shutdown after timeout")
	}
}
