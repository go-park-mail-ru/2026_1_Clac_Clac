package interceptors_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/interceptors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestPrometheusUnaryInterceptor(t *testing.T) {
	interceptor := interceptors.PrometheusUnaryInterceptor()
	require.NotNil(t, interceptor)

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	t.Run("successful handler", func(t *testing.T) {
		handler := func(ctx context.Context, req any) (any, error) {
			return "response", nil
		}

		resp, err := interceptor(context.Background(), nil, info, handler)

		assert.NoError(t, err)
		assert.Equal(t, "response", resp)
	})

	t.Run("handler returns error", func(t *testing.T) {
		wantErr := errors.New("handler error")
		handler := func(ctx context.Context, req any) (any, error) {
			return nil, wantErr
		}

		resp, err := interceptor(context.Background(), nil, info, handler)

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, resp)
	})

	t.Run("handler receives context and request", func(t *testing.T) {
		type ctxKey struct{}
		ctx := context.WithValue(context.Background(), ctxKey{}, "value")
		wantReq := "my-request"

		handler := func(gotCtx context.Context, gotReq any) (any, error) {
			assert.Equal(t, "value", gotCtx.Value(ctxKey{}))
			assert.Equal(t, wantReq, gotReq)
			return nil, nil
		}

		_, _ = interceptor(ctx, wantReq, info, handler)
	})
}

func TestPrometheusStreamInterceptor(t *testing.T) {
	interceptor := interceptors.PrometheusStreamInterceptor()
	require.NotNil(t, interceptor)

	info := &grpc.StreamServerInfo{FullMethod: "/test.Service/StreamMethod"}

	t.Run("successful handler", func(t *testing.T) {
		handler := func(srv any, stream grpc.ServerStream) error {
			return nil
		}

		err := interceptor(nil, nil, info, handler)
		assert.NoError(t, err)
	})

	t.Run("handler returns error", func(t *testing.T) {
		wantErr := errors.New("stream error")
		handler := func(srv any, stream grpc.ServerStream) error {
			return wantErr
		}

		err := interceptor(nil, nil, info, handler)
		assert.ErrorIs(t, err, wantErr)
	})
}
