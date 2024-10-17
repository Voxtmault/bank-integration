package bank_integration

import (
	"github.com/rotisserie/eris"
	"github.com/voxtmault/bank-integration/config"
	"github.com/voxtmault/bank-integration/storage"
)

func InitBankAPI(envPath string) error {
	// Load Configs
	cfg := config.New(envPath)

	// Init storage connections
	if err := storage.InitMariaDB(&cfg.MariaConfig); err != nil {
		return eris.Wrap(err, "init mariadb connection")
	}
	if err := storage.InitRedis(&cfg.RedisConfig); err != nil {
		return eris.Wrap(err, "init redis connection")
	}

	// Load Authenticated Banks to Redis

	// Load Services

	return nil
}
