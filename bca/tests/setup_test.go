package bca_test

import (
	"log/slog"
	"strings"

	bcaSecurity "github.com/voxtmault/bank-integration/bca/security"
	biConfig "github.com/voxtmault/bank-integration/config"
	biStorage "github.com/voxtmault/bank-integration/storage"
	biUtil "github.com/voxtmault/bank-integration/utils"
)

var envPath = "../../.env"
var bankEnvPath = "../../bca-mock.env"

var cfg *biConfig.InternalConfig
var bCfg *biConfig.BankConfig

func setup() (*bcaSecurity.BCASecurity, error) {
	cfg = biConfig.New(envPath)
	bCfg = biConfig.NewBankingConfig(bankEnvPath)

	validate := biUtil.InitValidator()
	validate.RegisterValidation("bcaPartnerServiceID", biUtil.ValidatePartnerServiceID)
	validate.RegisterValidation("bcaVA", biUtil.ValidateBCAVirtualAccountNumber)

	biStorage.InitMariaDB(&cfg.MariaConfig)
	biStorage.InitRedis(&cfg.RedisConfig)

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

func bcaMockSetup(bankPath, cfgPath string) (*bcaSecurity.BCASecurity, error) {
	cfg = biConfig.New(cfgPath)
	bCfg = biConfig.NewBankingConfig(bankPath)

	validate := biUtil.InitValidator()
	validate.RegisterValidation("bcaPartnerServiceID", biUtil.ValidatePartnerServiceID)
	validate.RegisterValidation("bcaVA", biUtil.ValidateBCAVirtualAccountNumber)

	biStorage.InitMariaDB(&cfg.MariaConfig)
	biStorage.InitRedis(&cfg.RedisConfig)

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
