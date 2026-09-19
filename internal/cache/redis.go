package cache

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jekyulll/url_shortener/config"
	"github.com/jekyulll/url_shortener/internal/model"
	"github.com/redis/go-redis/v9"
)

const urlPrefix = "url:"

type RedisCache struct {
	client            *redis.Client
	urlDuration       time.Duration
	emailCodeDuration time.Duration
	bloomFilterName   string
	bloomErrorRate    float64
	bloomCapacity     uint
	UseBloom          bool
	CacheTTL          time.Duration
}

func NewRedisCache(cfg config.RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	cache := &RedisCache{
		client:            client,
		bloomFilterName:   cfg.BloomFilterName,
		bloomErrorRate:    cfg.BloomErrorRate,
		bloomCapacity:     cfg.BloomCapacity,
		UseBloom:          true,
		CacheTTL:          cfg.CacheTTL,
		urlDuration:       cfg.URLDuration,
		emailCodeDuration: cfg.EmailCodeDuration,
	}
	// 初始化布隆过滤器
	err := cache.InitBloomFilter(context.Background(), cfg.BloomFilterName, cfg.BloomErrorRate, cfg.BloomCapacity)
	if err != nil {
		// 初始化错误的时候（可能没装这个模块），不退出
		cache.UseBloom = false
		log.Printf("failed to init redis bloom filter: %v", err)
	}
	return cache, nil
}

func (cache *RedisCache) InitBloomFilter(ctx context.Context, name string, errorRate float64, capacity uint) error {
	// 先检测布隆过滤器是否存在，简单做法是直接调用 BF.RESERVE 可能会报错
	// RedisBloom 并没有直接判断是否存在的命令，所以这里用 BF.RESERVE，失败了就忽略错误
	cmd := cache.client.Do(ctx, "BF.RESERVE", name, errorRate, capacity)
	err := cmd.Err()
	if err != nil {
		// 如果报错且包含“exists”字样，说明过滤器已存在，可以忽略
		if !strings.Contains(err.Error(), "exists") {
			return err
		}
	}
	return nil
}

func (cache *RedisCache) SetURL(ctx context.Context, url model.URL) error {
	if err := cache.BloomAdd(ctx, url.ShortCode); err != nil {
		// TODO 如果添加失败，之后再查找该短链接是否会直接显示不存在？是否需要处理？
		log.Printf("failed to set bloom filter: %v", err)
	}
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}
	// 如果配置里设置了缓存过期时间（大于0），就使用配置。否则直接设置为URL的过期时间
	var expiration time.Duration
	if cache.CacheTTL > 0 {
		expiration = cache.CacheTTL
	}
	expiration = time.Until(url.ExpiredAt)
	cmd := cache.client.Set(ctx, url.ShortCode, data, expiration)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

func (cache *RedisCache) GetURL(ctx context.Context, shortCode string) (*model.URL, error) {
	// 查找布隆过滤器
	exist, err := cache.BloomExists(ctx, shortCode)
	if err != nil {
		log.Printf("failed to read bloom filter: %v", err)
	}
	if !exist {
		return nil, nil
	}
	cmd := cache.client.Get(ctx, shortCode)
	if err := cmd.Err(); err != nil {
		// Get 不到值时会返回 redis.Nil，这是「缓存未命中」的正常情况
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	data := cmd.Val()
	var u model.URL
	if err := json.Unmarshal([]byte(data), &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (cache *RedisCache) DelURL(ctx context.Context, shortCode string) error {
	return cache.client.Del(ctx, urlPrefix+shortCode).Err()
}

func (cache *RedisCache) Close() error {
	return cache.client.Close()
}

func (cache *RedisCache) BloomAdd(ctx context.Context, item string) error {
	return cache.client.Do(ctx, "BF.ADD", cache.bloomFilterName, item).Err()
}

func (cache *RedisCache) BloomExists(ctx context.Context, item string) (bool, error) {
	val, err := cache.client.Do(ctx, "BF.EXISTS", cache.bloomFilterName, item).Int()
	return val == 1, err
}

// 加锁
func (cache *RedisCache) AcquireLock(ctx context.Context, lockKey string, expiration time.Duration) (string, bool, error) {
	lockValue := uuid.New().String() // 生成唯一值，避免误删别人的锁
	ok, err := cache.client.SetNX(ctx, lockKey, lockValue, expiration).Result()
	if err != nil {
		return "", false, err
	}
	return lockValue, ok, nil
}

// 解锁（通过 Lua 脚本保证只释放自己的锁）
func (cache *RedisCache) ReleaseLock(ctx context.Context, lockKey, lockValue string) (bool, error) {
	luaScript := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end`
	res, err := cache.client.Eval(ctx, luaScript, []string{lockKey}, lockValue).Int()
	return res == 1, err
}
