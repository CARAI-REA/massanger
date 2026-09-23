package interceptor

import (
	"context"
	"errors"
	"strings"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/tokens"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"rooms/internal/service/authctx"
)

const authorizationHeader = "authorization"

var (
	errMissingMetadata      = errors.New("missing metadata")
	errMissingAuthorization = errors.New("missing authorization")
	errInvalidAuthorization = errors.New("invalid authorization header")
)

var publicMethods = map[string]struct{}{
	"/grpc.health.v1.Health/Check": {},
	"/grpc.health.v1.Health/Watch": {},
	"/grpc.reflection.v1.ServerReflection/ServerReflectionInfo":      {},
	"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo": {},
}

// serviceMethods accept internal service JWT (signaling/sfu), not user access JWT.
var serviceMethods = map[string]struct{}{
	"/rooms.v1.RoomService/AssertCanJoin":       {},
	"/rooms.v1.RoomService/GetTURNCredentials": {},
}

// AuthInterceptor checks Bearer JWT: service JWT for AssertCanJoin/GetTURNCredentials,
// user access JWT for all other RPCs (including GetRoom, IsParticipant, RefreshJoinToken).
func AuthInterceptor(accessVerifier tokens.AccessTokenVerifier, serviceVerifier tokens.ServiceTokenVerifier) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if _, ok := publicMethods[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		token, err := bearerTokenFromContext(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		if _, ok := serviceMethods[info.FullMethod]; ok {
			claims, err := serviceVerifier.VerifyServiceToken(ctx, token)
			if err != nil {
				return nil, status.Error(codes.Unauthenticated, "invalid service token")
			}
			ctx = authctx.WithServiceName(ctx, claims.Service)
			return handler(ctx, req)
		}

		claims, err := accessVerifier.VerifyAccessToken(ctx, token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid access token")
		}

		ctx = authctx.WithUserID(ctx, claims.UserUUID)
		ctx = logger.ContextWithUserID(ctx, claims.UserUUID)
		return handler(ctx, req)
	}
}

func bearerTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errMissingMetadata
	}

	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return "", errMissingAuthorization
	}

	parts := strings.SplitN(strings.TrimSpace(values[0]), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", errInvalidAuthorization
	}

	return parts[1], nil
}
