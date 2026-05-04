package interceptors

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	grpcMetric "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/metrics/grpc"
)

func PrometheusUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		duration := time.Since(start).Seconds()

		st, _ := status.FromError(err)
		statusCode := st.Code().String()

		grpcMetric.GrpcRequestDuration.WithLabelValues(info.FullMethod).Observe(duration)
		grpcMetric.GrpcRequestsTotal.WithLabelValues(info.FullMethod, statusCode).Inc()

		return resp, err
	}
}

func PrometheusStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		err := handler(srv, ss)

		duration := time.Since(start).Seconds()

		st, _ := status.FromError(err)
		statusCode := st.Code().String()

		grpcMetric.GrpcRequestDuration.WithLabelValues(info.FullMethod).Observe(duration)
		grpcMetric.GrpcRequestsTotal.WithLabelValues(info.FullMethod, statusCode).Inc()

		return err
	}
}
