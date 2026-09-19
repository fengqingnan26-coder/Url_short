package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const viewPrifix = "views:"

func (r *RedisCache) IncreViews(ctx context.Context, shortCode string) error {
	return r.client.Incr(context.Background(), viewPrifix+shortCode).Err()
}

func (r *RedisCache) ScanViews(ctx context.Context, cursor uint64, batchSize int64) (keys []string, nextCursor uint64, err error) {
	return r.client.Scan(ctx, cursor, viewPrifix, batchSize).Result()
}

func (r *RedisCache) GetViews(ctx context.Context, shortCode string) (int, error) {
	views, err := r.client.Get(ctx, viewPrifix+shortCode).Int()
	if err == redis.Nil {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return views, nil
}

func (r *RedisCache) DelViews(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
