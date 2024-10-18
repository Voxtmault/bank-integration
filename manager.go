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
	obj, err := storage.InitRedis(&cfg.RedisConfig)
	if err != nil {
		return eris.Wrap(err, "init redis connection")
	}

	// Load Authenticated Banks to Redis
	if err := LoadAuthenticatedBanks(storage.GetDBConnection(), obj); err != nil {
		return eris.Wrap(err, "load authenticated banks")
	}

	// Load Services

	return nil
}
