package cache

import "context"

const emailPrifix = "email:"

func (cache *RedisCache) GetEmailCode(ctx context.Context, email string) (string, error) {
	emailCode := cache.client.Get(ctx, emailPrifix+email).Val()
	return emailCode, nil
}

func (cache *RedisCache) SetEmailCode(ctx context.Context, email, emailCode string) error {
	return cache.client.Set(ctx, emailPrifix+email, emailCode, cache.emailCodeDuration).Err()
}
