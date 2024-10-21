package bank_integration

import (
	"github.com/rotisserie/eris"
	bcaRequest "github.com/voxtmault/bank-integration/bca/request"
	bcaSecurity "github.com/voxtmault/bank-integration/bca/security"
	bcaService "github.com/voxtmault/bank-integration/bca/service"
	"github.com/voxtmault/bank-integration/config"
	"github.com/voxtmault/bank-integration/interfaces"
	"github.com/voxtmault/bank-integration/storage"
	"github.com/voxtmault/bank-integration/utils"
)

func InitBankAPI(envPath string) error {
	// Load Configs
	cfg := config.New(envPath)
	utils.InitValidator()

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

	return nil
}

func InitBCAService() interfaces.SNAP {

	security := bcaSecurity.NewBCASecurity(config.GetConfig())

	service := bcaService.NewBCAService(
		bcaRequest.NewBCAEgress(security),
		bcaRequest.NewBCAIngress(security),
		config.GetConfig(),
		storage.GetDBConnection(),
		storage.GetRedisInstance(),
	)

	return service
}
