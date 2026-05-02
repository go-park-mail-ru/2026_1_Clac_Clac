package interceptors

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func UnaryAccessLog(logger *zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		startTime := time.Now()

		requestID := getRequestID(ctx)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		requestLogger := logger.With().
			Str("method", info.FullMethod).
			Str("request_id", requestID).
			Logger()

		resp, err := handler(ctx, req)

		duration := time.Since(startTime)

		var logEvent *zerolog.Event
		grpcStatus := status.Convert(err)
		if grpcStatus.Code() == codes.OK {
			logEvent = requestLogger.Info()
		} else {
			logEvent = requestLogger.Error()
		}

		peerAddr := getPeerAddr(ctx)

		logEvent = logEvent.
			Str("url", info.FullMethod).
			Dur("work_time", duration).
			Int("status", int(grpcStatus.Code())).
			Str("start_time", startTime.Format(time.RFC3339)).
			Str("duration_human", duration.String()).
			Int64("duration_ms", duration.Milliseconds()).
			Str("remote_addr", peerAddr)

		logEvent.Msg("request processed")

		return resp, err
	}
}

func StreamAccessLog(logger *zerolog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startTime := time.Now()

		requestID := getRequestID(ss.Context())
		if requestID == "" {
			requestID = uuid.New().String()
		}

		requestLogger := logger.With().
			Str("method", info.FullMethod).
			Str("request_id", requestID).
			Logger()

		err := handler(srv, ss)

		duration := time.Since(startTime)

		var logEvent *zerolog.Event
		grpcStatus := status.Convert(err)
		if grpcStatus.Code() == codes.OK {
			logEvent = requestLogger.Info()
		} else {
			logEvent = requestLogger.Error()
		}

		peerAddr := getPeerAddr(ss.Context())

		logEvent = logEvent.
			Str("url", info.FullMethod).
			Dur("work_time", duration).
			Int("status", int(grpcStatus.Code())).
			Str("start_time", startTime.Format(time.RFC3339)).
			Str("duration_human", duration.String()).
			Int64("duration_ms", duration.Milliseconds()).
			Str("remote_addr", peerAddr)

		logEvent.Msg("request processed")

		return err
	}
}

func getRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md["x-request-id"]
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func getPeerAddr(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}
	return p.Addr.String()
}
