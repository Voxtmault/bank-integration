package bank_integration

import (
	"testing"

	biConfig "github.com/voxtmault/bank-integration/config"
	biStorage "github.com/voxtmault/bank-integration/storage"
)

var envPath = "./.env"
var bankPath = "./bca.env"

func TestInitBCAService(t *testing.T) {

	InitBankAPI(envPath, "Asia/Jakarta")

	cfg := biConfig.New(envPath)
	biStorage.InitMariaDB(&cfg.MariaConfig)
	biStorage.InitRedis(&cfg.RedisConfig)

	_, err := InitBCAService(bankPath)
	if err != nil {
		t.Fatalf("error initializing bca service: %v", err)
	}
}
