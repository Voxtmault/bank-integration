package bank_integration

import (
	"context"
	"log/slog"
	"time"

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

	// Run background job to clear unique external id every day after midnight
	go func() {
		for {
			now := time.Now()
			next := now.AddDate(0, 0, 1).Truncate(24 * time.Hour)
			time.Sleep(time.Until(next))

			if err := clearList(context.Background(), obj, "unique_external_id:*"); err != nil {
				slog.Info("failed to clear unique external id", "reason", err)
			} else {
				slog.Info("unique external id cleared")
			}
		}
	}()

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

func clearList(ctx context.Context, rdb *storage.RedisInstance, pattern string) error {
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
