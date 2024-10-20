package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Keys struct {
	PrivateKeyPath   string
	BCAPublicKeyPath string
}

// Banking Config
type BCAConfig struct {
	BaseURL      string `validate:"required"`
	ClientID     string `validate:"required"`
	ClientSecret string `validate:"required"`
	AccessToken  string `validate:"omitempty"`
}

type BCAPartnerId struct {
	BCAPartnerId string `validate:"required"`
}

type BCAURLEndpoints struct {
	AccessTokenURL    string
	BalanceInquiryURL string
}

// Credential DB Config
type MariaConfig struct {
	DBDriver             string
	DBHost               string
	DBPort               string
	DBUser               string
	DBName               string
	DBPassword           string
	TSLConfig            string
	AllowNativePasswords bool
	MultiStatements      bool
	MaxOpenConns         uint
	MaxIdleConns         uint
	ConnMaxLifetime      uint
}
type RedisConfig struct {
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDBNum    uint8
}

type BankingConfig struct {
	Keys
	BCAConfig
	BCAURLEndpoints
	MariaConfig
	RedisConfig
	BCAPartnerId
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
			PrivateKeyPath:   getEnv("PRIVATE_KEY_PATH", ""),
			BCAPublicKeyPath: getEnv("BCA_PUBLIC_KEY_PATH", ""),
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
		BCAPartnerId: BCAPartnerId{
			BCAPartnerId: getEnv("BCA_PARTNER_ID", ""),
		},
		MariaConfig: MariaConfig{
			DBDriver:             getEnv("DB_DRIVER", "mysql"),
			DBHost:               getEnv("DB_HOST", ""),
			DBPort:               getEnv("DB_PORT", "3306"),
			DBUser:               getEnv("DB_USER", ""),
			DBPassword:           getEnv("DB_PASSWORD", ""),
			DBName:               getEnv("DB_NAME", ""),
			TSLConfig:            getEnv("DB_TLS_CONFIG", "true"),
			AllowNativePasswords: getEnvAsBool("DB_ALLOW_NATIVE_PASSWORDS", true),
			MultiStatements:      getEnvAsBool("DB_MULTI_STATEMENTS", false),
			MaxOpenConns:         uint(getEnvAsInt("DB_MAX_OPEN_CONNS", 20)),
			MaxIdleConns:         uint(getEnvAsInt("DB_MAX_IDLE_CONNS", 5)),
			ConnMaxLifetime:      uint(getEnvAsInt("DB_CONN_MAX_LIFETIME", 5)),
		},
		RedisConfig: RedisConfig{
			RedisHost:     getEnv("REDIS_HOST", ""),
			RedisPort:     getEnv("REDIS_PORT", "6378"),
			RedisPassword: getEnv("REDIS_PASSWORD", ""),
			RedisDBNum:    uint8(getEnvAsInt("REDIS_DB_NUM", 0)),
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
// func getEnvAsSlice[T any](name string, defaultVal []T, sep string) []T {
// 	valStr := getEnv(name, "")

// 	if valStr == "" {
// 		return defaultVal
// 	}

// 	vals := strings.Split(valStr, sep)
// 	result := make([]T, len(vals))

// 	for i, v := range vals {
// 		switch any(result).(type) {
// 		case []string:
// 			result[i] = any(v).(T)
// 		case []int:
// 			intVal, _ := strconv.Atoi(v)
// 			result[i] = any(intVal).(T)
// 		case []bool:
// 			boolVal, _ := strconv.ParseBool(v)
// 			result[i] = any(boolVal).(T)
// 		default:
// 			return defaultVal
// 		}
// 	}

// 	return result
// }
