package interceptors

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryPanicRecovery(logger *zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				requestID := getRequestID(ctx)
				logger.Error().
					Str("method", info.FullMethod).
					Str("request_id", requestID).
					Str("panic", fmt.Sprintf("%v", r)).
					Msg("panic recovered")

				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

func StreamPanicRecovery(logger *zerolog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		defer func() {
			if r := recover(); r != nil {
				requestID := getRequestID(ss.Context())
				logger.Error().
					Str("method", info.FullMethod).
					Str("request_id", requestID).
					Str("panic", fmt.Sprintf("%v", r)).
					Msg("panic recovered")
			}
		}()

		return handler(srv, ss)
	}
}
