package redisclient

import (
	"context"
	"encoding/json"
	"fmt"

	"eventum/internal/event"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

func New(ctx context.Context, addr string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	return &RedisClient{
		client: rdb,
		ctx:    ctx,
	}
}

func (r *RedisClient) SaveEvent(key string, ev *event.Event) error {
	data, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = r.client.Set(r.ctx, key, string(data), 0).Err()
	if err != nil {
		return fmt.Errorf("failed to save event in redis: %w", err)
	}

	return nil
}

func (r *RedisClient) GetEvent(key string) (*event.Event, error) {
	data, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("event not found for key: %s", key)
	} else if err != nil {
		return nil, fmt.Errorf("failed to retrieve event from redis: %w", err)
	}

	var ev event.Event
	err = json.Unmarshal([]byte(data), &ev)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	return &ev, nil
}

func (r *RedisClient) DeleteEvent(key string) error {
	err := r.client.Del(r.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete event from redis: %w", err)
	}

	return nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
