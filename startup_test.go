package bank_integration

import (
	"log/slog"
	"testing"

	"github.com/voxtmault/bank-integration/config"
	"github.com/voxtmault/bank-integration/storage"
)

func TestLoadAuthenticatedBanks(t *testing.T) {
	cfg := config.New("/home/andy/go-projects/github.com/voxtmault/bank-integration/.env")
	if err := storage.InitMariaDB(&cfg.MariaConfig); err != nil {
		t.Errorf("init mariadb: %v", err)
	}
	redis, err := storage.InitRedis(&cfg.RedisConfig)
	if err != nil {
		t.Errorf("init redis: %v", err)
	}

	slog.SetLogLoggerLevel(slog.LevelDebug)

	if err := LoadAuthenticatedBanks(storage.GetDBConnection(), redis); err != nil {
		t.Errorf("load authenticated banks: %v", err)
	}
}
