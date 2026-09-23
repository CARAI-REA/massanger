package interceptor

import (
	"context"
	"testing"
	"time"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/tokens"
	jwtTokens "github.com/CARAI-REA/messanger/callService/platform/pkg/tokens/jwt"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"rooms/internal/service/authctx"
)

type stubAccessVerifier struct {
	claims *tokens.AccessClaims
	err    error
}

func (s stubAccessVerifier) VerifyAccessToken(context.Context, string) (*tokens.AccessClaims, error) {
	return s.claims, s.err
}

type stubServiceVerifier struct {
	claims *tokens.ServiceClaims
	err    error
}

func (s stubServiceVerifier) VerifyServiceToken(context.Context, string) (*tokens.ServiceClaims, error) {
	return s.claims, s.err
}

func TestAuthInterceptorSuccess(t *testing.T) {
	userUUID := gofakeit.UUID()
	interceptor := AuthInterceptor(stubAccessVerifier{
		claims: &tokens.AccessClaims{UserUUID: userUUID},
	}, stubServiceVerifier{})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		authorizationHeader, "Bearer test-token",
	))

	var gotUserID string
	_, err := interceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/rooms.v1.RoomService/CreateRoom"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			id, ok := authctx.UserID(ctx)
			require.True(t, ok)
			gotUserID = id
			return "ok", nil
		},
	)

	require.NoError(t, err)
	require.Equal(t, userUUID, gotUserID)
}

func TestAuthInterceptorMissingToken(t *testing.T) {
	interceptor := AuthInterceptor(stubAccessVerifier{}, stubServiceVerifier{})

	_, err := interceptor(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/rooms.v1.RoomService/CreateRoom"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			t.Fatal("handler must not be called")
			return nil, nil
		},
	)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, st.Code())
}

func TestAuthInterceptorInvalidToken(t *testing.T) {
	interceptor := AuthInterceptor(stubAccessVerifier{err: jwt.ErrTokenMalformed}, stubServiceVerifier{})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		authorizationHeader, "Bearer bad-token",
	))

	_, err := interceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/rooms.v1.RoomService/CreateRoom"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			t.Fatal("handler must not be called")
			return nil, nil
		},
	)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, st.Code())
}

func TestAuthInterceptorSkipsHealth(t *testing.T) {
	interceptor := AuthInterceptor(stubAccessVerifier{err: jwt.ErrTokenMalformed}, stubServiceVerifier{})

	called := false
	_, err := interceptor(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/grpc.health.v1.Health/Check"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			called = true
			return "ok", nil
		},
	)

	require.NoError(t, err)
	require.True(t, called)
}

func TestAuthInterceptorRequiresAuthForGetRoomAndIsParticipant(t *testing.T) {
	interceptor := AuthInterceptor(stubAccessVerifier{err: jwt.ErrTokenMalformed}, stubServiceVerifier{})

	for _, method := range []string{
		"/rooms.v1.RoomService/GetRoom",
		"/rooms.v1.RoomService/IsParticipant",
	} {
		_, err := interceptor(
			context.Background(),
			nil,
			&grpc.UnaryServerInfo{FullMethod: method},
			func(ctx context.Context, req interface{}) (interface{}, error) {
				t.Fatal("handler must not be called")
				return nil, nil
			},
		)
		require.Error(t, err, method)
		st, ok := status.FromError(err)
		require.True(t, ok, method)
		require.Equal(t, codes.Unauthenticated, st.Code(), method)
	}
}

func TestAuthInterceptorServiceJWTForAssertCanJoin(t *testing.T) {
	interceptor := AuthInterceptor(
		stubAccessVerifier{err: jwt.ErrTokenMalformed},
		stubServiceVerifier{claims: &tokens.ServiceClaims{Service: "signaling"}},
	)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		authorizationHeader, "Bearer service-token",
	))

	var gotService string
	_, err := interceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/rooms.v1.RoomService/AssertCanJoin"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			name, ok := authctx.ServiceName(ctx)
			require.True(t, ok)
			gotService = name
			return "ok", nil
		},
	)

	require.NoError(t, err)
	require.Equal(t, "signaling", gotService)
}

func TestAuthInterceptorServiceJWTForGetTURNCredentials(t *testing.T) {
	interceptor := AuthInterceptor(
		stubAccessVerifier{err: jwt.ErrTokenMalformed},
		stubServiceVerifier{claims: &tokens.ServiceClaims{Service: "sfu"}},
	)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		authorizationHeader, "Bearer service-token",
	))

	called := false
	_, err := interceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/rooms.v1.RoomService/GetTURNCredentials"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			called = true
			name, ok := authctx.ServiceName(ctx)
			require.True(t, ok)
			require.Equal(t, "sfu", name)
			return "ok", nil
		},
	)

	require.NoError(t, err)
	require.True(t, called)
}

func TestAuthInterceptorRejectsUserJWTOnServiceMethod(t *testing.T) {
	interceptor := AuthInterceptor(
		stubAccessVerifier{claims: &tokens.AccessClaims{UserUUID: gofakeit.UUID()}},
		stubServiceVerifier{err: jwt.ErrTokenMalformed},
	)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		authorizationHeader, "Bearer user-token",
	))

	_, err := interceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/rooms.v1.RoomService/AssertCanJoin"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			t.Fatal("handler must not be called")
			return nil, nil
		},
	)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, st.Code())
}

func TestAuthInterceptorRefreshJoinTokenRequiresUserJWT(t *testing.T) {
	userUUID := gofakeit.UUID()
	interceptor := AuthInterceptor(
		stubAccessVerifier{claims: &tokens.AccessClaims{UserUUID: userUUID}},
		stubServiceVerifier{err: jwt.ErrTokenMalformed},
	)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		authorizationHeader, "Bearer user-token",
	))

	var gotUserID string
	_, err := interceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/rooms.v1.RoomService/RefreshJoinToken"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			id, ok := authctx.UserID(ctx)
			require.True(t, ok)
			gotUserID = id
			return "ok", nil
		},
	)

	require.NoError(t, err)
	require.Equal(t, userUUID, gotUserID)
}

func TestAccessJWTVerifierRoundTrip(t *testing.T) {
	secret := "test-auth-secret"
	userUUID := gofakeit.UUID()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokens.AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserUUID: userUUID,
	})
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	verifier := accessVerifierForTest(secret)
	claims, err := verifier.VerifyAccessToken(context.Background(), signed)
	require.NoError(t, err)
	require.Equal(t, userUUID, claims.UserUUID)
}

type secretCfg string

func (s secretCfg) AuthTokenSecretKey() string { return string(s) }

func accessVerifierForTest(secret string) tokens.AccessTokenVerifier {
	return jwtTokens.NewAccessJWTVerifier(secretCfg(secret))
}
