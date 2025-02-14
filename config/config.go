package bank_integration_config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type BankCredential struct {
	InternalBankID   uint   `validate:"omitempty,number,gte=1,min=1"` // Refers to internally registered bank
	InternalBankName string `validate:"omitempty"`                    // Name of the bank
	ClientID         string `validate:"required,uuid4"`               // Client ID received from the bank
	ClientSecret     string `validate:"required,uuid4"`               // Client Secret received from the bank
	VAPrefix         string `validate:"required"`                     // Virtual Account Prefix
	PartnerID        string `validate:"required,max=32"`              // Partner ID / Company ID / Corporate ID received from the bank
	PublicKeyPath    string `validate:"required,filepath"`            // Path to the public key sent by the bank
	SourceAccount    string `validate:"required"`                     // Source account numbern for this application
}

type BankChannelConfig struct {
	VAChannelId       string `validate:"required,number"` // Virtual Account Channel ID
	BusinessChannelId string `validate:"required,number"` // Business Channel ID; KlikBCA Bisnis, etc
}

type BankRuntimeConfig struct {
	AccessToken               string // Access token received from the bank on runtime
	AccessTokenExpirationTime uint   // Access token expiration time, this can be used to determine when to refresh the token
	ExpiresAt                 int64
}

type BankRequestedCredentials struct {
	ClientID              string `validate:"required,uuid4"`              // Client ID given to the bank
	ClientSecret          string `validate:"required,uuid4"`              // Client Secret given to the bank
	AccessTokenExpireTime uint   `validate:"required,number,min=1,gte=1"` // Access token expiration time, this can be used to determine when to recycle the token
}

type RequestedEndpoints struct {
	AuthURL            string `validate:"required,uri"`
	BillPresentmentURL string `validate:"required,uri"`
	PaymentFlagURL     string `validate:"required,uri"`
}

// Used to store the endpoints of the bank's API for each specific operation
type BankServiceEndpoints struct {
	BaseUrl                   string `validate:"required,url"` // Base URL to the bank's API
	AccessTokenURL            string `validate:"required,uri"` // URL to get the access token
	BalanceInquiryURL         string `validate:"required,uri"` // URL to check / get the balance information of an account
	PaymentFlagURL            string `validate:"required,uri"` // URL to update / play a billing statement
	TransferIntraBankURL      string `validate:"required,uri"` // URL to transfer money / withdraw money from the application to the target account (only for intra bank)
	TransferInterBankURL      string `validate:"required,uri"` // URL to transfer money / withdraw money from the application to the target account (only for inter bank)
	ExternalAccountInquiryURL string `validate:"required,uri"` // URL to check / get the information of an external (non-bca) account
	InternalAccountInquiryURL string `validate:"required,uri"` // URL to check / get the information of an internal (bca) account
	BankStatementURL          string `validate:"required,uri"` // URL to check / get the information of a billing statement
}

type VirtualAccountConfig struct {
	VirtualAccountLife uint `validate:"required,number,min=1,gte=1"`
}

// BankConfig is used to store / bundle configuration needed to run / create a bank instance
type BankConfig struct {
	BankCredential
	BankChannelConfig
	BankRuntimeConfig
	BankRequestedCredentials
	BankServiceEndpoints
	RequestedEndpoints
	VirtualAccountConfig
}

func NewBankingConfig(path string) *BankConfig {
	if err := godotenv.Load(path); err != nil {
		log.Println("Failed to locate .env file, program will proceed with provided env if any is provided")
	}

	return &BankConfig{
		BankCredential: BankCredential{
			InternalBankID:   uint(getEnvAsInt("INTERNAL_BANK_ID", 0)),
			InternalBankName: getEnv("INTERNAL_BANK_NAME", ""),
			ClientID:         getEnv("CLIENT_ID", ""),
			ClientSecret:     getEnv("CLIENT_SECRET", ""),
			VAPrefix:         getEnv("VA_PREFIX", ""),
			PartnerID:        getEnv("PARTNER_ID", ""),
			PublicKeyPath:    getEnv("PUBLIC_KEY_PATH", ""),
			SourceAccount:    getEnv("SOURCE_ACCOUNT", ""),
		},
		BankChannelConfig: BankChannelConfig{
			VAChannelId:       getEnv("VA_CHANNEL_ID", ""),
			BusinessChannelId: getEnv("BUSINESS_CHANNEL_ID", ""),
		},
		BankRuntimeConfig: BankRuntimeConfig{
			AccessTokenExpirationTime: uint(getEnvAsInt("ACCESS_TOKEN_EXPIRATION_TIME", 0)),
		},
		BankRequestedCredentials: BankRequestedCredentials{
			ClientID:              getEnv("REQ_CLIENT_ID", ""),
			ClientSecret:          getEnv("REQ_CLIENT_SECRET", ""),
			AccessTokenExpireTime: uint(getEnvAsInt("REQ_ACCESS_TOKEN_EXPIRATION", 0)),
		},
		BankServiceEndpoints: BankServiceEndpoints{
			BaseUrl:                   getEnv("BASE_URL", ""),
			AccessTokenURL:            getEnv("ACCESS_TOKEN_URL", ""),
			BalanceInquiryURL:         getEnv("BALANCE_INQUIRY_URL", ""),
			PaymentFlagURL:            getEnv("BANK_PAYMENT_FLAG_URL", ""),
			TransferIntraBankURL:      getEnv("TRANSFER_INTRABANK_URL", ""),
			TransferInterBankURL:      getEnv("TRANSFER_INTERBANK_URL", ""),
			ExternalAccountInquiryURL: getEnv("EXTERNAL_ACCOUNT_INQUIRY_URL", ""),
			InternalAccountInquiryURL: getEnv("INTERNAL_ACCOUNT_INQUIRY_URL", ""),
			BankStatementURL:          getEnv("BANK_STATEMENT_URL", ""),
		},
		RequestedEndpoints: RequestedEndpoints{
			AuthURL:            getEnv("OAUTH2_URL", ""),
			BillPresentmentURL: getEnv("BILL_PRESENTMENT_URL", ""),
			PaymentFlagURL:     getEnv("PAYMENT_FLAG_URL", ""),
		},
		VirtualAccountConfig: VirtualAccountConfig{
			VirtualAccountLife: uint(getEnvAsInt("VIRTUAL_ACCOUNT_LIFE", 24)),
		},
	}
}

// For internal use

type TransactionWatcherConfig struct {
	MaxRetry             uint
	DefaultRetryInterval time.Duration
	DefaultExpireTime    time.Duration
}

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

type ForwardProxyConfig struct {
	ProxyAddress string
}

type InternalConfig struct {
	TransactionWatcherConfig
	MariaConfig
	RedisConfig
	ForwardProxyConfig
	PrivateKeyPath string
	AppHost        string
	TZ             string
	Mode           string // To contorl wether the application is running in production or development or debug mode
}

var config *InternalConfig

func New(envPath string) *InternalConfig {

	if err := godotenv.Load(envPath); err != nil {
		log.Println("Failed to locate .env file, program will proceed with provided env if any is provided")
	}

	config = &InternalConfig{
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
		ForwardProxyConfig: ForwardProxyConfig{
			ProxyAddress: getEnv("PROXY_ADDRESS", ""),
		},
		TransactionWatcherConfig: TransactionWatcherConfig{
			MaxRetry:             uint(getEnvAsInt("WATCHER_MAX_RETRY", 10)),
			DefaultRetryInterval: time.Duration(getEnvAsInt("WATCHER_DEFAULT_RETRY_INTERVAL", 10)) * time.Minute,
			DefaultExpireTime:    time.Duration(getEnvAsInt("WATCHER_DEFAULT_EXPIRE_TIME", 24)) * time.Hour,
		},
		PrivateKeyPath: getEnv("PRIVATE_KEY_PATH", ""),
		AppHost:        getEnv("APP_HOST", ""),
		Mode:           getEnv("MODE", "prod"),
		TZ:             getEnv("TZ", "Asia/Jakarta"),
	}

	return config
}

func GetConfig() *InternalConfig {
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
