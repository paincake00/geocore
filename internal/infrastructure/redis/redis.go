package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/paincake00/geocore/internal/entity"
	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client *redis.Client
}

func New(addr string) (*RedisRepo, error) {
	// In a real app we might want password/db options too.
	// For now assuming simple host:port string
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisRepo{Client: client}, nil
}

func (r *RedisRepo) Close() {
	r.Client.Close()
}

func (r *RedisRepo) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

// Queue

func (r *RedisRepo) Enqueue(ctx context.Context, queueName string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return r.Client.LPush(ctx, queueName, data).Err()
}

func (r *RedisRepo) Dequeue(ctx context.Context, queueName string) (string, error) {
	// BRPOP blocks until an item is available. timeout 0 means block indefinitely.
	result, err := r.Client.BRPop(ctx, 0, queueName).Result()
	if err != nil {
		return "", err
	}
	// result is [queue, value]
	if len(result) < 2 {
		return "", fmt.Errorf("redis pop unexpected result")
	}
	return result[1], nil
}

// Cache

const IncidentsCacheKey = "active_incidents"

func (r *RedisRepo) SetIncidents(ctx context.Context, incidents []*entity.Incident) error {
	data, err := json.Marshal(incidents)
	if err != nil {
		return err
	}
	// TTL can be 1 minute or less, or we can rely on manual invalidation.
	// Given "Service synchronously returns", we probably want fast read.
	// Let's set 60s TTL for safety, or indefinite if we handle updates.
	// For simplicity, 1 minute TTL seems fine for verification.
	return r.Client.Set(ctx, IncidentsCacheKey, data, 60*time.Second).Err()
}

func (r *RedisRepo) GetIncidents(ctx context.Context) ([]*entity.Incident, error) {
	val, err := r.Client.Get(ctx, IncidentsCacheKey).Result()
	if err == redis.Nil {
		return nil, nil // cache miss
	}
	if err != nil {
		return nil, err
	}

	var incidents []*entity.Incident
	if err := json.Unmarshal([]byte(val), &incidents); err != nil {
		return nil, err
	}
	return incidents, nil
}
