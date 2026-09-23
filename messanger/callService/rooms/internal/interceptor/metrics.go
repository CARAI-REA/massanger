package interceptor

import (
	"context"
	"path"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"rooms/internal/metrics"
)

func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		method := path.Base(info.FullMethod)
		start := time.Now()

		resp, err := handler(ctx, req)

		code := status.Code(err).String()
		metrics.RequestsTotal.WithLabelValues(method, code).Inc()
		metrics.RequestDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())

		return resp, err
	}
}
