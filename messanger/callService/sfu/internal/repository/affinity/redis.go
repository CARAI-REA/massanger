package affinity

import (
	"context"
	"fmt"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	"sfu/internal/metrics"
)

type Redis struct {
	pool *redigo.Pool
}

func NewRedis(pool *redigo.Pool) *Redis {
	return &Redis{pool: pool}
}

func (r *Redis) withConn(ctx context.Context, fn func(redigo.Conn) error) error {
	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	return fn(conn)
}

func (r *Redis) Claim(ctx context.Context, roomUUID, instanceID, publicURL string, ttl time.Duration) (bool, string, error) {
	key := ownerKey(roomUUID)
	val := encodeOwner(instanceID, publicURL)
	var claimed bool
	var ownerURL string
	err := r.withConn(ctx, func(conn redigo.Conn) error {
		reply, err := redigo.String(conn.Do("SET", key, val, "NX", "EX", int(ttl.Seconds())))
		if err == redigo.ErrNil || reply == "" {
			cur, err := redigo.String(conn.Do("GET", key))
			if err == redigo.ErrNil {
				return nil
			}
			if err != nil {
				return err
			}
			id, url := decodeOwner(cur)
			if id == instanceID {
				_, _ = conn.Do("EXPIRE", key, int(ttl.Seconds()))
				claimed = true
				ownerURL = url
				return nil
			}
			ownerURL = url
			return nil
		}
		if err != nil {
			return err
		}
		claimed = true
		ownerURL = publicURL
		return nil
	})
	if err != nil {
		metrics.RedisErrorsTotal.WithLabelValues("claim").Inc()
		return false, "", err
	}
	return claimed, ownerURL, nil
}

func (r *Redis) Refresh(ctx context.Context, roomUUID, instanceID string, ttl time.Duration) error {
	err := r.withConn(ctx, func(conn redigo.Conn) error {
		cur, err := redigo.String(conn.Do("GET", ownerKey(roomUUID)))
		if err == redigo.ErrNil {
			return nil
		}
		if err != nil {
			return err
		}
		id, _ := decodeOwner(cur)
		if id != instanceID {
			return fmt.Errorf("not owner")
		}
		_, err = conn.Do("EXPIRE", ownerKey(roomUUID), int(ttl.Seconds()))
		return err
	})
	if err != nil {
		metrics.RedisErrorsTotal.WithLabelValues("refresh").Inc()
	}
	return err
}

func (r *Redis) Release(ctx context.Context, roomUUID, instanceID string) error {
	err := r.withConn(ctx, func(conn redigo.Conn) error {
		cur, err := redigo.String(conn.Do("GET", ownerKey(roomUUID)))
		if err == redigo.ErrNil {
			return nil
		}
		if err != nil {
			return err
		}
		id, _ := decodeOwner(cur)
		if id != instanceID {
			return nil
		}
		_, err = conn.Do("DEL", ownerKey(roomUUID))
		return err
	})
	if err != nil {
		metrics.RedisErrorsTotal.WithLabelValues("release").Inc()
	}
	return err
}
