package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type (
	Cache interface {
		Get(ctx context.Context, key string) ([]byte, error)
		Put(ctx context.Context, key string, val []byte) error
		Invalidate(ctx context.Context, key string) (bool, error)
		PutWithTTL(ctx context.Context, key string, val []byte, ttl time.Duration) error
	}

	Redis struct {
		r *redis.Client
	}
)

var _ Cache = (*Redis)(nil)

func NewRedis(ctx context.Context, addr, password string) (*Redis, error) {
	r := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	if _, err := r.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &Redis{r: r}, nil
}

func (s *Redis) Get(ctx context.Context, key string) ([]byte, error) {
	sc := s.r.Get(ctx, key)
	if sc.Err() != nil {
		return nil, sc.Err()
	}

	return sc.Bytes()
}

func (s *Redis) Put(ctx context.Context, key string, val []byte) error {
	sc := s.r.Set(ctx, key, val, 0)
	return sc.Err()
}

func (s *Redis) PutWithTTL(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	sc := s.r.Set(ctx, key, val, ttl)
	return sc.Err()
}

func (s *Redis) Invalidate(ctx context.Context, key string) (bool, error) {
	sc := s.r.Del(ctx, key)
	return sc.Val() > 0, sc.Err()
}

func (s *Redis) Close() error {
	return s.r.Close()
}
