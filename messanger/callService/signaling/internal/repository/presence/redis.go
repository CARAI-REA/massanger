package presence

import (
	"context"
	"fmt"
	"sync"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	"signaling/internal/converter"
	"signaling/internal/metrics"
	"signaling/internal/model"
)

type Redis struct {
	pool *redigo.Pool
	ttl  time.Duration

	mu       sync.Mutex
	refcount map[string]int
	subs     map[string]*redigo.PubSubConn
	handler  func(roomUUID string, msg model.Envelope)
	cancel   context.CancelFunc
}

func NewRedis(pool *redigo.Pool, ttl time.Duration) *Redis {
	return &Redis{
		pool:     pool,
		ttl:      ttl,
		refcount: make(map[string]int),
		subs:     make(map[string]*redigo.PubSubConn),
	}
}

func (r *Redis) withConn(ctx context.Context, fn func(redigo.Conn) error) error {
	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	return fn(conn)
}

func (r *Redis) Register(ctx context.Context, roomUUID, userUUID, instanceID string) error {
	err := r.withConn(ctx, func(conn redigo.Conn) error {
		if err := conn.Send("SADD", peersKey(roomUUID), userUUID); err != nil {
			return err
		}
		if err := conn.Send("HSET", peerKey(roomUUID, userUUID),
			"instance_id", instanceID,
			"connected_at", time.Now().Unix(),
		); err != nil {
			return err
		}
		if err := conn.Send("EXPIRE", peerKey(roomUUID, userUUID), int(r.ttl.Seconds())); err != nil {
			return err
		}
		return conn.Flush()
	})
	if err != nil {
		metrics.RedisErrorsTotal.WithLabelValues("register").Inc()
	}
	return err
}

func (r *Redis) Unregister(ctx context.Context, roomUUID, userUUID string) error {
	err := r.withConn(ctx, func(conn redigo.Conn) error {
		if err := conn.Send("SREM", peersKey(roomUUID), userUUID); err != nil {
			return err
		}
		if err := conn.Send("DEL", peerKey(roomUUID, userUUID)); err != nil {
			return err
		}
		return conn.Flush()
	})
	if err != nil {
		metrics.RedisErrorsTotal.WithLabelValues("unregister").Inc()
	}
	return err
}

func (r *Redis) RefreshTTL(ctx context.Context, roomUUID, userUUID string) error {
	err := r.withConn(ctx, func(conn redigo.Conn) error {
		_, err := conn.Do("EXPIRE", peerKey(roomUUID, userUUID), int(r.ttl.Seconds()))
		return err
	})
	if err != nil {
		metrics.RedisErrorsTotal.WithLabelValues("refresh").Inc()
	}
	return err
}

func (r *Redis) ListPeers(ctx context.Context, roomUUID string) ([]string, error) {
	var peers []string
	err := r.withConn(ctx, func(conn redigo.Conn) error {
		vals, err := redigo.Strings(conn.Do("SMEMBERS", peersKey(roomUUID)))
		if err != nil {
			return err
		}
		peers = vals
		return nil
	})
	if err != nil {
		metrics.RedisErrorsTotal.WithLabelValues("list").Inc()
		return nil, err
	}
	return peers, nil
}

func (r *Redis) GetInstance(ctx context.Context, roomUUID, userUUID string) (string, bool, error) {
	var id string
	err := r.withConn(ctx, func(conn redigo.Conn) error {
		val, err := redigo.String(conn.Do("HGET", peerKey(roomUUID, userUUID), "instance_id"))
		if err == redigo.ErrNil {
			return nil
		}
		if err != nil {
			return err
		}
		id = val
		return nil
	})
	if err != nil {
		metrics.RedisErrorsTotal.WithLabelValues("get_instance").Inc()
		return "", false, err
	}
	if id == "" {
		return "", false, nil
	}
	return id, true, nil
}

func (r *Redis) Publish(ctx context.Context, roomUUID string, msg model.Envelope) error {
	body, err := converter.MarshalEnvelope(msg)
	if err != nil {
		return err
	}
	err = r.withConn(ctx, func(conn redigo.Conn) error {
		_, err := conn.Do("PUBLISH", pubsubChannel(roomUUID), body)
		return err
	})
	if err != nil {
		metrics.RedisErrorsTotal.WithLabelValues("publish").Inc()
	}
	return err
}

func (r *Redis) SetHandler(handler func(roomUUID string, msg model.Envelope)) {
	r.mu.Lock()
	r.handler = handler
	r.mu.Unlock()
}

func (r *Redis) Subscribe(ctx context.Context, handler func(roomUUID string, msg model.Envelope)) error {
	if handler != nil {
		r.SetHandler(handler)
	}
	r.mu.Lock()
	_, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	r.mu.Unlock()

	<-ctx.Done()
	r.mu.Lock()
	for _, psc := range r.subs {
		_ = psc.Unsubscribe()
		_ = psc.Close()
	}
	r.subs = make(map[string]*redigo.PubSubConn)
	r.refcount = make(map[string]int)
	r.mu.Unlock()
	return ctx.Err()
}

func (r *Redis) EnsureSubscribed(ctx context.Context, roomUUID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refcount[roomUUID]++
	if r.refcount[roomUUID] > 1 {
		return nil
	}
	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		r.refcount[roomUUID]--
		metrics.RedisErrorsTotal.WithLabelValues("subscribe").Inc()
		return err
	}
	psc := &redigo.PubSubConn{Conn: conn}
	if err := psc.Subscribe(pubsubChannel(roomUUID)); err != nil {
		_ = conn.Close()
		r.refcount[roomUUID]--
		metrics.RedisErrorsTotal.WithLabelValues("subscribe").Inc()
		return err
	}
	r.subs[roomUUID] = psc
	go r.readLoop(roomUUID, psc)
	return nil
}

func (r *Redis) ReleaseSubscribe(_ context.Context, roomUUID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.refcount[roomUUID] <= 0 {
		return nil
	}
	r.refcount[roomUUID]--
	if r.refcount[roomUUID] > 0 {
		return nil
	}
	if psc, ok := r.subs[roomUUID]; ok {
		_ = psc.Unsubscribe()
		_ = psc.Close()
		delete(r.subs, roomUUID)
	}
	delete(r.refcount, roomUUID)
	return nil
}

func (r *Redis) readLoop(roomUUID string, psc *redigo.PubSubConn) {
	for {
		switch v := psc.Receive().(type) {
		case redigo.Message:
			msg, err := converter.UnmarshalEnvelope(v.Data)
			if err != nil {
				continue
			}
			r.mu.Lock()
			h := r.handler
			r.mu.Unlock()
			if h != nil {
				h(roomUUID, msg)
			}
		case redigo.Subscription:
			if v.Count == 0 {
				return
			}
		case error:
			return
		}
	}
}

// Ping checks redis connectivity.
func (r *Redis) Ping(ctx context.Context) error {
	return r.withConn(ctx, func(conn redigo.Conn) error {
		pong, err := redigo.String(conn.Do("PING"))
		if err != nil {
			return err
		}
		if pong != "PONG" {
			return fmt.Errorf("unexpected ping reply: %s", pong)
		}
		return nil
	})
}
