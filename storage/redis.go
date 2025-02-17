package bank_integration_storage

import (
	"context"
	"fmt"
	"log/slog"

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
