package interceptor

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"rooms/internal/service/authctx"
)

type roomIDGetter interface {
	GetRoomUuid() string
}

func LoggerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		method := path.Base(info.FullMethod)
		requestID := uuid.NewString()
		ctx = logger.ContextWithTraceID(ctx, requestID)

		if userID, ok := authctx.UserID(ctx); ok {
			ctx = logger.ContextWithUserID(ctx, userID)
		}

		fields := []zap.Field{
			zap.String("service", "rooms"),
			zap.String("method", method),
			zap.String("request_id", requestID),
		}
		if roomID := roomIDFromRequest(req); roomID != "" {
			fields = append(fields, zap.String("room_id", roomID))
		}

		logger.Info(ctx, "Started gRPC method", fields...)

		startTime := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(startTime)

		fields = append(fields, zap.Duration("duration", duration))
		if err != nil {
			st, _ := status.FromError(err)
			fields = append(fields, zap.String("code", st.Code().String()), zap.Error(err))
			logger.Error(ctx, fmt.Sprintf("Finished gRPC method %s", method), fields...)
		} else {
			fields = append(fields, zap.String("code", "OK"))
			logger.Info(ctx, fmt.Sprintf("Finished gRPC method %s", method), fields...)
		}

		return resp, err
	}
}

func roomIDFromRequest(req interface{}) string {
	if g, ok := req.(roomIDGetter); ok {
		return g.GetRoomUuid()
	}
	return ""
}
