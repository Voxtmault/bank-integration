package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Keys struct {
	PrivateKeyPath string
	PublicKeyPath  string
}

// Banking Config
type BCAConfig struct {
	BaseURL      string `validate:"required"`
	ClientID     string `validate:"required"`
	ClientSecret string `validate:"required"`
	AccessToken  string `validate:"omitempty"`
}

type BCAURLEndpoints struct {
	AccessTokenURL    string
	BalanceInquiryURL string
}

type BankingConfig struct {
	Keys
	BCAConfig
	BCAURLEndpoints
	AppHost string
	Mode    string
}

var config *BankingConfig

func New(envPath string) *BankingConfig {

	if err := godotenv.Load(envPath); err != nil {
		log.Println("Failed to locate .env file, program will proceed with provided env if any is provided")
	}

	config = &BankingConfig{
		Keys: Keys{
			PrivateKeyPath: getEnv("PRIVATE_KEY_PATH", ""),
			PublicKeyPath:  getEnv("PUBLIC_KEY_PATH", ""),
		},
		BCAConfig: BCAConfig{
			BaseURL:      getEnv("BCA_BASE_URL", ""),
			ClientID:     getEnv("BCA_CLIENT_ID", ""),
			ClientSecret: getEnv("BCA_CLIENT_SECRET", ""),
		},
		BCAURLEndpoints: BCAURLEndpoints{
			AccessTokenURL:    getEnv("BCA_ACCESS_TOKEN_URL", ""),
			BalanceInquiryURL: getEnv("BCA_BALANCE_INQUIRY_URL", ""),
		},
		AppHost: getEnv("APP_HOST", ""),
		Mode:    getEnv("MODE", "prod"),
	}

	return config
}

func GetConfig() *BankingConfig {
	return config
}

// Simple helper function to read an environment or return a default value.
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	if nextValue := os.Getenv(key); nextValue != "" {
		return nextValue
	}

	return defaultVal
}

// Simple helper function to read an environment variable into integer or return a default value.
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

// Helper to read an environment variable into a bool or return default value.
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

// Helper to read an environment variable into a slice of a specific type or return default value.
func getEnvAsSlice[T any](name string, defaultVal []T, sep string) []T {
	valStr := getEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	vals := strings.Split(valStr, sep)
	result := make([]T, len(vals))

	for i, v := range vals {
		switch any(result).(type) {
		case []string:
			result[i] = any(v).(T)
		case []int:
			intVal, _ := strconv.Atoi(v)
			result[i] = any(intVal).(T)
		case []bool:
			boolVal, _ := strconv.ParseBool(v)
			result[i] = any(boolVal).(T)
		default:
			return defaultVal
		}
	}

	return result
}
