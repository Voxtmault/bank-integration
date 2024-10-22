package bank_integration

import (
	"log/slog"
	"testing"

	biConfig "github.com/voxtmault/bank-integration/config"
	biStorage "github.com/voxtmault/bank-integration/storage"
)

func TestLoadAuthenticatedBanks(t *testing.T) {
	cfg := biConfig.New("/home/andy/go-projects/github.com/voxtmault/bank-integration/.env")
	if err := biStorage.InitMariaDB(&cfg.MariaConfig); err != nil {
		t.Errorf("init mariadb: %v", err)
	}
	redis, err := biStorage.InitRedis(&cfg.RedisConfig)
	if err != nil {
		t.Errorf("init redis: %v", err)
	}

	slog.SetLogLoggerLevel(slog.LevelDebug)

	if err := LoadAuthenticatedBanks(biStorage.GetDBConnection(), redis); err != nil {
		t.Errorf("load authenticated banks: %v", err)
	}
}
