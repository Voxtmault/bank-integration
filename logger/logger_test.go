package bank_integration_logger_test

import (
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	biConfig "github.com/voxtmault/bank-integration/config"
	biLogger "github.com/voxtmault/bank-integration/logger"
	bank_integration_models "github.com/voxtmault/bank-integration/models"
	biStorage "github.com/voxtmault/bank-integration/storage"
	biUtil "github.com/voxtmault/bank-integration/utils"
)

var envPath = "../.env"

var cfg *biConfig.InternalConfig

func setup() error {
	cfg = biConfig.New(envPath)

	biUtil.InitValidator()

	biStorage.InitMariaDB(&cfg.MariaConfig, &cfg.LoggerConfig)
	biStorage.InitRedis(&cfg.RedisConfig)
	biLogger.InitLogger()

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	return nil
}

func TestLogger(t *testing.T) {
	slog.Debug("setting up")
	err := setup()
	if err != nil {
		t.Errorf("Failed to setup: %v", err)
	}

	slog.Debug("done setting up")

	biLogger.LogRequest(&bank_integration_models.BankLogV2{
		IDBank:         100,
		BeginAt:        time.Now(),
		EndAt:          time.Now().Add(time.Millisecond * 150),
		IDFeature:      100,
		ResponseCode:   http.StatusOK,
		ClientIP:       "192.168.1.1",
		HTTPMethod:     http.MethodPost,
		Protocol:       "HTTP/1.1",
		URI:            "/api/v1/bank",
		RequestHeader:  "{}",
		RequestBody:    "{}",
		ResponseHeader: "{}",
		ResponseBody:   "{}",
	})

	time.Sleep(time.Second * 2)
}
