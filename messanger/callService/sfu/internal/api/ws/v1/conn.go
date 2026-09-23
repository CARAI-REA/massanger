package v1

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"sfu/internal/converter"
	"sfu/internal/metrics"
	"sfu/internal/model"
	"sfu/internal/service"
)

type ClientConn struct {
	ws         *websocket.Conn
	send       chan model.Envelope
	roomUUID   string
	userUUID   string
	pingEvery  time.Duration
	pongWait   time.Duration
	maxBytes   int64
	limiter    *rateLimiter
	sessions   service.SessionService
	remoteAddr string
	writeMu    sync.Mutex
	closeOnce  sync.Once
	closed     chan struct{}
}

func NewClientConn(
	ws *websocket.Conn,
	roomUUID, userUUID string,
	pingEvery, pongWait time.Duration,
	maxBytes int64,
	maxMsgPerSec int,
	sessions service.SessionService,
) *ClientConn {
	return &ClientConn{
		ws:         ws,
		send:       make(chan model.Envelope, 64),
		roomUUID:   roomUUID,
		userUUID:   userUUID,
		pingEvery:  pingEvery,
		pongWait:   pongWait,
		maxBytes:   maxBytes,
		limiter:    newRateLimiter(maxMsgPerSec),
		sessions:   sessions,
		remoteAddr: ws.RemoteAddr().String(),
		closed:     make(chan struct{}),
	}
}

func (c *ClientConn) Send(_ context.Context, msg model.Envelope) error {
	select {
	case <-c.closed:
		return net.ErrClosed
	default:
	}
	select {
	case <-c.closed:
		return net.ErrClosed
	case c.send <- msg:
		return nil
	default:
		return net.ErrClosed
	}
}

func (c *ClientConn) SendAndClose(_ context.Context, msg model.Envelope, code int, reason string) error {
	_ = c.writeEnvelope(msg)
	return c.Close(code, reason)
}

func (c *ClientConn) Close(code int, reason string) error {
	var err error
	c.closeOnce.Do(func() {
		close(c.closed)
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			select {
			case msg := <-c.send:
				_ = c.writeEnvelope(msg)
			default:
				goto flushed
			}
		}
	flushed:
		c.writeMu.Lock()
		_ = c.ws.SetWriteDeadline(time.Now().Add(time.Second))
		_ = c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason))
		err = c.ws.Close()
		c.writeMu.Unlock()
	})
	return err
}

func (c *ClientConn) RemoteAddr() string { return c.remoteAddr }

func (c *ClientConn) writeEnvelope(msg model.Envelope) error {
	body, err := converter.MarshalEnvelope(msg)
	if err != nil {
		return err
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_ = c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.ws.WriteMessage(websocket.TextMessage, body)
}

func (c *ClientConn) writePump(ctx context.Context) {
	ticker := time.NewTicker(c.pingEvery)
	defer ticker.Stop()
	for {
		select {
		case <-c.closed:
			return
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.writeEnvelope(msg); err != nil {
				return
			}
		case <-ticker.C:
			c.writeMu.Lock()
			_ = c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := c.ws.WriteMessage(websocket.PingMessage, nil)
			c.writeMu.Unlock()
			if err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *ClientConn) readPump(ctx context.Context) {
	defer func() {
		_ = c.sessions.Disconnect(ctx, c, c.roomUUID, c.userUUID, "disconnect")
	}()
	c.ws.SetReadLimit(c.maxBytes)
	_ = c.ws.SetReadDeadline(time.Now().Add(c.pongWait))
	c.ws.SetPongHandler(func(string) error {
		_ = c.ws.SetReadDeadline(time.Now().Add(c.pongWait))
		return nil
	})
	for {
		_, data, err := c.ws.ReadMessage()
		if err != nil {
			return
		}
		_ = c.ws.SetReadDeadline(time.Now().Add(c.pongWait))
		if !c.limiter.Allow() {
			metrics.MessageErrorsTotal.WithLabelValues(model.CodeRateLimited).Inc()
			_ = c.Send(ctx, converter.ErrorEnvelope(c.roomUUID, model.CodeRateLimited, "rate limited"))
			if c.limiter.Violations() >= 3 {
				return
			}
			continue
		}
		msg, err := converter.UnmarshalEnvelope(data)
		if err != nil {
			_ = c.Send(ctx, converter.ErrorEnvelope(c.roomUUID, model.CodeInvalidArgument, "invalid json"))
			continue
		}
		if err := c.sessions.Handle(ctx, c.roomUUID, c.userUUID, msg); err != nil {
			logger.Error(ctx, "handle message failed", zap.Error(err))
		}
	}
}

func (c *ClientConn) Run(ctx context.Context) {
	go c.writePump(ctx)
	c.readPump(ctx)
}
