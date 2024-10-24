package bank_integration

import (
	"context"
	"testing"

	biConfig "github.com/voxtmault/bank-integration/config"
	biStorage "github.com/voxtmault/bank-integration/storage"
)

var envPath = "./.env"

func TestClearList(t *testing.T) {
	cfg := biConfig.New(envPath)
	rdb, err := biStorage.InitRedis(&cfg.RedisConfig)
	if err != nil {
		t.Error(err)
	}

	pattern := "unique-external-id:*"

	var cursor uint64
	for {
		keys, nextCursor, err := rdb.RDB.Scan(context.Background(), cursor, pattern, 100).Result()
		if err != nil {
			t.Error(err)
		}

		if len(keys) > 0 {
			if err := rdb.RDB.Del(context.Background(), keys...).Err(); err != nil {
				t.Error(err)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
}
