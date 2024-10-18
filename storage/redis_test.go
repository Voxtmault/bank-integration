package storage

import (
	"testing"

	"github.com/voxtmault/bank-integration/config"
)

func TestInitRedis(t *testing.T) {
	cfg := config.New("/home/andy/go-projects/github.com/voxtmault/bank-integration/.env")
	obj, err := InitRedis(&cfg.RedisConfig)
	if err != nil {
		t.Errorf("Error initializing redis: %v", err)
	}

	if err := obj.CloseRedis(); err != nil {
		t.Errorf("Error closing redis: %v", err)
	}
}
