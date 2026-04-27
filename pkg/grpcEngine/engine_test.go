package grpcEngine

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestGracefulShutdown(t *testing.T) {
	const timeout = 16 * time.Second
	cfg := Config{
		Addr:                    ":0",
		WriteTimeout:            15,
		ReadTimeout:             15,
		IdleTimeout:             60,
		GracefulShutdownTimeout: 15,
	}

	buf := &bytes.Buffer{}
	logger := zerolog.New(buf)

	e := New(cfg, &logger)

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
