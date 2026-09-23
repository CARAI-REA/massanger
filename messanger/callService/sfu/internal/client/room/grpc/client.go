package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/tokens"
	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"sfu/internal/model"
)

type Client struct {
	conn        *grpc.ClientConn
	client      roomsV1.RoomServiceClient
	tokenGen    tokens.ServiceTokenGenerator
	serviceName string
}

func New(addr string, tokenGen tokens.ServiceTokenGenerator, serviceName string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func (c *Client) AssertCanJoin(ctx context.Context, roomUUID, userUUID string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	ctx, err := c.withServiceAuth(ctx)
	if err != nil {
		return err
	}

	resp, err := c.client.AssertCanJoin(ctx, &roomsV1.AssertCanJoinRequest{
		RoomUuid: roomUUID,
		UserUuid: userUUID,
	})
	if err != nil {
		return fmt.Errorf("%w: assert can join: %v", model.ErrPermissionDenied, err)
	}
	if resp.GetOk() {
		return nil
	}
	status := resp.GetRoomStatus()
	if status != "" && status != "active" {
		return model.ErrRoomClosed
	}
	return model.ErrPermissionDenied
}
