package bca_test

import (
	"log/slog"
	"strings"

	bcaSecurity "github.com/voxtmault/bank-integration/bca/security"
	biConfig "github.com/voxtmault/bank-integration/config"
	biLogger "github.com/voxtmault/bank-integration/logger"
	biStorage "github.com/voxtmault/bank-integration/storage"
	biUtil "github.com/voxtmault/bank-integration/utils"
)

var envPath = "../../.env"
var bankEnvPath = "../../bca-testing.env"

var cfg *biConfig.InternalConfig
var bCfg *biConfig.BankConfig

func setup() (*bcaSecurity.BCASecurity, error) {
	cfg = biConfig.New(envPath)
	bCfg = biConfig.NewBankingConfig(bankEnvPath)

	validate := biUtil.InitValidator()
	validate.RegisterValidation("bcaPartnerServiceID", biUtil.ValidatePartnerServiceID)
	validate.RegisterValidation("bcaVA", biUtil.ValidateBCAVirtualAccountNumber)

	biStorage.InitMariaDB(&cfg.MariaConfig, &cfg.LoggerConfig)
	biStorage.InitRedis(&cfg.RedisConfig)
	biLogger.InitLogger()

	security, err := bcaSecurity.NewBCASecurity(cfg, bCfg)
	if err != nil {
		return nil, err
	}

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	return security, nil
}
