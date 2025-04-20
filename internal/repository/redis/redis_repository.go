package redis

import (
	"GoTasker/internal/config"
	"GoTasker/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

type AnalyticsRedisRepo struct {
	client *redis.Client
	cfg    *config.Config
}

func NewAnalyticsRedisRepo(cfg *config.Config) *AnalyticsRedisRepo {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	return &AnalyticsRedisRepo{
		client: client,
		cfg:    cfg,
	}
}

func (r *AnalyticsRedisRepo) GetAnalytics(ctx context.Context) (*domain.AnalyticsTasksResponse, error) {
	const op = "internal.repository.redis.GetAnalytics"

	val, err := r.client.Get(ctx, "analytics_cache").Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	} else if err != nil {
		slog.Error(op, "ошибка получения данных из Redis", slog.String("err", err.Error()))
		return nil, fmt.Errorf("ошибка получения данных из Redis: %w", err)
	}

	var analytics domain.AnalyticsTasksResponse
	if err = json.Unmarshal([]byte(val), &analytics); err != nil {
		slog.Error(op, "ошибка десериализации данных", slog.String("err", err.Error()))
		return nil, fmt.Errorf("ошибка десериализации данных: %w", err)
	}

	return &analytics, nil
}

func (r *AnalyticsRedisRepo) SetAnalytics(ctx context.Context, analytics *domain.AnalyticsTasksResponse) error {
	const op = "internal.repository.redis.SetAnalytics"

	data, err := json.Marshal(analytics)
	if err != nil {
		slog.Error(op, "ошибка сериализации данных", slog.String("err", err.Error()))
		return fmt.Errorf("ошибка сериализации данных: %w", err)
	}

	err = r.client.Set(ctx, "analytics_cache", string(data), r.cfg.Redis.TTL).Err()
	if err != nil {
		slog.Error(op, "ошибка сохранения данных в Redis", slog.String("err", err.Error()))
		return fmt.Errorf("ошибка сохранения данных в Redis: %w", err)
	}

	return nil
}
