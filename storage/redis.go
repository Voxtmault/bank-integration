package bank_integration_storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rotisserie/eris"
	biConfig "github.com/voxtmault/bank-integration/config"
)

type RedisInstance struct {
	RDB *redis.Client
}

var redisInstance RedisInstance

func validateRedisConfig(cfg *biConfig.RedisConfig) error {
	if cfg.RedisHost == "" {
		return eris.New("redis host is empty")
	}
	if cfg.RedisPort == "" {
		return eris.New("redis port is empty")
	}
	if cfg.RedisPassword == "" {
		return eris.New("redis password is empty")
	}

	return nil
}

func InitRedis(config *biConfig.RedisConfig) (*RedisInstance, error) {

	slog.Debug("Validating Redis Config")
	if err := validateRedisConfig(config); err != nil {
		return nil, eris.Wrap(err, "invalid redis configuration")
	}

	redisInstance.RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
		Password: config.RedisPassword,
		DB:       int(config.RedisDBNum),
	})

	if _, err := redisInstance.RDB.Ping(context.Background()).Result(); err != nil {
		return nil, eris.Wrap(err, "Init Redis")
	}

	slog.Debug("Successfully opened redis connection")

	return &redisInstance, nil
}

func GetRedisInstance() *RedisInstance {
	return &redisInstance
}

func (r *RedisInstance) CloseRedis() error {
	if redisInstance.RDB != nil {
		if err := r.RDB.Close(); err != nil {
			return eris.Wrap(err, "Closing redis connection")
		}
		return nil
	} else {
		slog.Info("Redis connection is already closed or is not opened in the first place")
		return nil
	}
}

func (r *RedisInstance) SaveToRedis(ctx context.Context, key string, value interface{}, exp time.Duration) error {
	if err := r.RDB.Set(ctx, key, value, exp).Err(); err != nil {
		return eris.Wrap(err, "saving data to redis cache")
	}

	return nil
}

func (r *RedisInstance) SaveRedisHash(ctx context.Context, key string, value map[string]interface{}) error {

	if err := r.RDB.HSet(ctx, key, value).Err(); err != nil {
		slog.Debug("error saving hash data to redis cache", "error", err)
		return eris.Wrap(err, "saving hash data to redis cache")
	}

	return nil
}

func (r *RedisInstance) GetIndividualValueRedisHash(ctx context.Context, key, subKey string) (string, error) {

	data, err := r.RDB.HGet(ctx, key, subKey).Result()
	if err != nil {
		if err == redis.Nil {
			slog.Debug("key does not exist in redis cache", "key", key)
			return "", nil
		} else {
			slog.Debug("error getting individual value hash data from redis cache", "error", err)
			return "", eris.Wrap(err, "getting individual value hash data from redis cache")
		}
	}

	return data, nil
}

func (r *RedisInstance) GetRedisHash(ctx context.Context, key string) (map[string]string, error) {

	data, err := r.RDB.HGetAll(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			slog.Debug("key does not exist in redis cache", "key", key)
			return nil, nil
		} else {
			slog.Debug("error getting hash data from redis cache", "error", err)
			return nil, eris.Wrap(err, "getting hash data from redis cache")
		}
	}

	return data, nil
}
