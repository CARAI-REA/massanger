package interceptor

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"rooms/internal/service/authctx"
)

// RateLimitInterceptor applies a per-key token bucket. maxRPS <= 0 disables limiting.
// Key is authctx user_id when present, otherwise peer address.
func RateLimitInterceptor(maxRPS float64) grpc.UnaryServerInterceptor {
	if maxRPS <= 0 {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}

	burst := int(maxRPS)
	if burst < 1 {
		burst = 1
	}

	var (
		mu       sync.Mutex
		limiters = make(map[string]*rate.Limiter)
	)

	getLimiter := func(key string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		if lim, ok := limiters[key]; ok {
			return lim
		}
		lim := rate.NewLimiter(rate.Limit(maxRPS), burst)
		limiters[key] = lim
		return lim
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		key := rateLimitKey(ctx)
		if !getLimiter(key).Allow() {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}
		return handler(ctx, req)
	}
}

func rateLimitKey(ctx context.Context) string {
	if userID, ok := authctx.UserID(ctx); ok {
		return "user:" + userID
	}
	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		return "peer:" + p.Addr.String()
	}
	return "peer:unknown"
}
