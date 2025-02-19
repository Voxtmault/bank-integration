package bank_integration

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rotisserie/eris"
	bcaRequest "github.com/voxtmault/bank-integration/bca/request"
	bcaSecurity "github.com/voxtmault/bank-integration/bca/security"
	bcaService "github.com/voxtmault/bank-integration/bca/service"
	biCache "github.com/voxtmault/bank-integration/cache"
	biConfig "github.com/voxtmault/bank-integration/config"
	biInterfaces "github.com/voxtmault/bank-integration/interfaces"
	bank_integration_internal "github.com/voxtmault/bank-integration/internal"
	management "github.com/voxtmault/bank-integration/management"
	biStorage "github.com/voxtmault/bank-integration/storage"
	biUtil "github.com/voxtmault/bank-integration/utils"
)

func InitBankAPI(envPath, timezone string) error {

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return eris.Wrap(err, "failed to load timezone")
	}

	// Load Configs
	cfg := biConfig.New(envPath)

	// Init Validator
	validate := biUtil.InitValidator()

	// Register custom validator func if any
	validate.RegisterValidation("bcaPartnerServiceID", biUtil.ValidatePartnerServiceID)
	validate.RegisterValidation("bcaVA", biUtil.ValidateBCAVirtualAccountNumber)

	// Init storage connections
	if err := biStorage.InitMariaDB(&cfg.MariaConfig, &cfg.LoggerConfig); err != nil {
		return eris.Wrap(err, "init mariadb connection")
	}
	obj, err := biStorage.InitRedis(&cfg.RedisConfig)
	if err != nil {
		return eris.Wrap(err, "init redis connection")
	}

	// Load Authenticated Banks to Redis
	// if err := LoadAuthenticatedBanks(biStorage.GetDBConnection(), obj); err != nil {
	// 	return eris.Wrap(err, "load authenticated banks")
	// }

	// Create a new cron scheduler
	c := cron.New(cron.WithLocation(location))

	// Schedule the task to run every day at midnight
	_, err = c.AddFunc("0 0 * * *", func() {
		if err := clearList(context.Background(), obj, "unique-external-id:*"); err != nil {
			slog.Info("failed to clear unique external id", "reason", err)
		} else {
			slog.Info("unique external id cleared")
		}
	})
	if err != nil {
		return eris.Wrap(err, "failed to schedule task")
	}

	_, err = c.AddFunc("* * * * *", func() {
		if err := biCache.InitCache(context.Background()); err != nil {
			slog.Info("failed to fetch helper data from redis", "reason", err)
		} else {
			slog.Debug("renewed cache data")
		}
	})
	if err != nil {
		return eris.Wrap(err, "failed to schedule task")
	}

	// Start the cron scheduler
	c.Start()

	return nil
}

func InitBCAService(envPath string) (biInterfaces.SNAP, error) {
	cfg := biConfig.NewBankingConfig(envPath)

	// Checks for problematic configurations
	if err := biUtil.ValidateStruct(context.Background(), cfg); err != nil {
		return nil, eris.Wrap(err, "invalid bank configuration")
	}

	security, err := bcaSecurity.NewBCASecurity(biConfig.GetConfig(), cfg)
	if err != nil {
		slog.Error("failed to init bca security instance", "reason", err)
		return nil, eris.Wrap(err, "init bca security")
	}

	service, err := bcaService.NewBCAService(
		bcaRequest.NewBCAEgress(security, cfg, biConfig.GetConfig()),
		bcaRequest.NewBCAIngress(security),
		biConfig.GetConfig(),
		cfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)

	return service, err
}

func GetBCAService() (biInterfaces.SNAP, error) {
	service := bcaService.GetBCAService()
	if service == nil {
		return nil, eris.New("bca service not initialized")
	}

	return service, nil
}

func InitManagementService() (biInterfaces.Management, error) {

	service, err := management.NewBankIntegrationManagement(
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)
	if err != nil {
		slog.Error("failed to init management service", "reason", err)
		return nil, eris.Wrap(err, "init management service")
	}

	return service, nil
}

func InitInternalService() biInterfaces.Internal {
	service, _ := bank_integration_internal.NewInternalService(
		biConfig.GetConfig(),
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)

	return service
}

func clearList(ctx context.Context, rdb *biStorage.RedisInstance, pattern string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := rdb.RDB.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := rdb.RDB.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

func CloseBankAPI() {
	if err := biStorage.Close(); err != nil {
		slog.Error("failed to close storage connections", "reason", err)
	}
	if err := biStorage.GetRedisInstance().CloseRedis(); err != nil {
		slog.Error("failed to close redis connection", "reason", err)
	}
}
