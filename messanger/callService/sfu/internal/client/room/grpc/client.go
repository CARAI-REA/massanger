package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/tokens"
	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"sfu/internal/client/room"
	"sfu/internal/model"
)

type Client struct {
	conn        *grpc.ClientConn
	client      roomsV1.RoomServiceClient
	tokenGen    tokens.ServiceTokenGenerator
	serviceName string
}

func New(addr string, tokenGen tokens.ServiceTokenGenerator, serviceName string) (*Client, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:        conn,
		client:      roomsV1.NewRoomServiceClient(conn),
		tokenGen:    tokenGen,
		serviceName: serviceName,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) withServiceAuth(ctx context.Context) (context.Context, error) {
	if c.tokenGen == nil {
		return ctx, nil
	}
	token, err := c.tokenGen.GenerateServiceToken(ctx, c.serviceName, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("service token: %w", err)
	}
	md := metadata.Pairs("authorization", "Bearer "+token)
	return metadata.NewOutgoingContext(ctx, md), nil
}

func (c *Client) AssertCanJoin(ctx context.Context, roomUUID, userUUID string) (room.JoinCheck, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	ctx, err := c.withServiceAuth(ctx)
	if err != nil {
		return room.JoinCheck{}, err
	}

	resp, err := c.client.AssertCanJoin(ctx, &roomsV1.AssertCanJoinRequest{
		RoomUuid: roomUUID,
		UserUuid: userUUID,
	})
	if err != nil {
		return room.JoinCheck{}, fmt.Errorf("%w: assert can join: %v", model.ErrPermissionDenied, err)
	}
	check := room.JoinCheck{
		OK:               resp.GetOk(),
		RoomStatus:       resp.GetRoomStatus(),
		RecordingEnabled: resp.GetRecordingEnabled(),
	}
	if check.OK {
		return check, nil
	}
	if check.RoomStatus != "" && check.RoomStatus != "active" {
		return check, model.ErrRoomClosed
	}
	return check, model.ErrPermissionDenied
}

func (c *Client) GetTURNCredentials(ctx context.Context, roomUUID, userUUID string) ([]model.ICEServer, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	ctx, err := c.withServiceAuth(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.GetTURNCredentials(ctx, &roomsV1.GetTURNCredentialsRequest{
		RoomUuid: roomUUID,
		UserUuid: userUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: get turn credentials: %v", model.ErrTURNUnavailable, err)
	}
	out := make([]model.ICEServer, 0, len(resp.GetIceServers()))
	for _, s := range resp.GetIceServers() {
		out = append(out, model.ICEServer{
			URLs:       append([]string(nil), s.GetUrls()...),
			Username:   s.GetUsername(),
			Credential: s.GetCredential(),
		})
	}
	return out, nil
}
