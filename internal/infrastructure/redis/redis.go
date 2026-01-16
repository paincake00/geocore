package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/paincake00/geocore/internal/entity"
	"github.com/redis/go-redis/v9"
)

// RedisRepo реализация репозитория на основе Redis (для очереди и кеша).
type RedisRepo struct {
	Client *redis.Client
}

// New создает новое подключение к Redis.
func New(addr string) (*RedisRepo, error) {
	// В реальном приложении стоило бы добавить настройки пароля и номера БД.
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisRepo{Client: client}, nil
}

// Close закрывает соединение.
func (r *RedisRepo) Close() {
	r.Client.Close()
}

// Ping проверяет доступность Redis.
func (r *RedisRepo) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

// Queue (Очередь)

// Enqueue добавляет задачу в очередь списка (LPush).
func (r *RedisRepo) Enqueue(ctx context.Context, queueName string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return r.Client.LPush(ctx, queueName, data).Err()
}

// Dequeue извлекает задачу из очереди (BRPop - блокирующее чтение).
func (r *RedisRepo) Dequeue(ctx context.Context, queueName string) (string, error) {
	// BRPop блокирует выполнение, пока не появится элемент. 0 - бесконечное ожидание.
	result, err := r.Client.BRPop(ctx, 0, queueName).Result()
	if err != nil {
		return "", err
	}
	// result содержит [имя_очереди, значение]
	if len(result) < 2 {
		return "", fmt.Errorf("redis pop unexpected result")
	}
	return result[1], nil
}

// Cache (Кеш)

const IncidentsCacheKey = "active_incidents"

// SetIncidents сохраняет список инцидентов в кеш с TTL.
func (r *RedisRepo) SetIncidents(ctx context.Context, incidents []*entity.Incident) error {
	data, err := json.Marshal(incidents)
	if err != nil {
		return err
	}
	// TTL настроен на 60 секунд.
	return r.Client.Set(ctx, IncidentsCacheKey, data, 60*time.Second).Err()
}

// GetIncidents получает список инцидентов из кеша.
func (r *RedisRepo) GetIncidents(ctx context.Context) ([]*entity.Incident, error) {
	val, err := r.Client.Get(ctx, IncidentsCacheKey).Result()
	if err == redis.Nil {
		return nil, nil // кеш пуст
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
