package bank_integration_storage

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/rotisserie/eris"
	biConfig "github.com/voxtmault/bank-integration/config"
)

var (
	mariaCon       *sql.DB
	loggerMariaCon *sql.DB
)

type MariaDatabaseStats struct {
	OpenConnections      int           `json:"open_connections"`
	ConnectionInUse      int           `json:"connection_in_use"`
	ConnectionIdle       int           `json:"connection_idle"`
	WaitingForConnection int           `json:"waiting_for_connection"`
	TotalWaitTime        time.Duration `json:"total_wait_time"`
}

func validateMariaDBConfig(config *biConfig.MariaConfig) error {
	if config.DBUser == "" {
		return eris.New("db username is empty")
	}
	if config.DBPassword == "" {
		return eris.New("db password is empty")
	}
	if config.DBHost == "" || config.DBPort == "" {
		return eris.New("invalid db address and or port")
	}
	if config.DBName == "" {
		return eris.New("invalid db name")
	}

	return nil
}

// InitMaria Establish a connection using the provided credentials with the mariadb service
func InitMariaDB(config *biConfig.MariaConfig, loggerConfig *biConfig.MariaConfig) error {
	slog.Debug("Opening MariaDB Connection")
	var err error

	// Validation
	slog.Debug("validating MariaDB Config")
	if err := validateMariaDBConfig(config); err != nil {
		return eris.Wrap(err, "invalid MariaDB configuration")
	}

	slog.Debug("validating logger MariaDB config")
	if err := validateMariaDBConfig(loggerConfig); err != nil {
		return eris.Wrap(err, "invalid logger MariaDB configuration")
	}

	dsn := mysql.Config{
		User:                 config.DBUser,
		Passwd:               config.DBPassword,
		AllowNativePasswords: config.AllowNativePasswords,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", config.DBHost, config.DBPort),
		DBName:               config.DBName,
		TLSConfig:            config.TSLConfig,
		MultiStatements:      config.MultiStatements,
		Params: map[string]string{
			"charset": "utf8",
		},
	}

	mariaCon, err = sql.Open(config.DBDriver, dsn.FormatDSN())
	if err != nil {
		return eris.Wrap(err, "Opening MySQL/MariaDB Connection")
	}

	mariaCon.SetMaxOpenConns(int(config.MaxOpenConns))
	mariaCon.SetMaxIdleConns(int(config.MaxIdleConns))
	mariaCon.SetConnMaxLifetime(time.Second * time.Duration(config.ConnMaxLifetime))

	err = mariaCon.Ping()
	if err != nil {
		return eris.Wrap(err, "Error verifying database connection")
	}

	loggerDsn := mysql.Config{
		User:                 loggerConfig.DBUser,
		Passwd:               loggerConfig.DBPassword,
		AllowNativePasswords: loggerConfig.AllowNativePasswords,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", loggerConfig.DBHost, loggerConfig.DBPort),
		DBName:               loggerConfig.DBName,
		TLSConfig:            loggerConfig.TSLConfig,
		MultiStatements:      loggerConfig.MultiStatements,
		Params: map[string]string{
			"charset": "utf8",
		},
	}

	loggerMariaCon, err = sql.Open(loggerConfig.DBDriver, loggerDsn.FormatDSN())
	if err != nil {
		return eris.Wrap(err, "Opening MySQL/MariaDB Connection")
	}

	loggerMariaCon.SetMaxOpenConns(int(loggerConfig.MaxOpenConns))
	loggerMariaCon.SetMaxIdleConns(int(loggerConfig.MaxIdleConns))
	loggerMariaCon.SetConnMaxLifetime(time.Second * time.Duration(loggerConfig.ConnMaxLifetime))

	err = loggerMariaCon.Ping()
	if err != nil {
		return eris.Wrap(err, "Error verifying database connection")
	}

	slog.Debug("Successfully opened database connection !")
	return nil
}

func GetDBConnection() *sql.DB {
	return mariaCon
}

func GetLoggerDBConnection() *sql.DB {
	return loggerMariaCon
}

// GetMariaStats
func GetDBStats() MariaDatabaseStats {
	return MariaDatabaseStats{
		OpenConnections:      mariaCon.Stats().OpenConnections,
		ConnectionInUse:      mariaCon.Stats().InUse,
		ConnectionIdle:       mariaCon.Stats().Idle,
		WaitingForConnection: int(mariaCon.Stats().WaitCount),
		TotalWaitTime:        mariaCon.Stats().WaitDuration,
	}
}

// CloseMaria will close the current database connection, only do this when exiting the program
//
// Under normal circumstances, this shouldn't be called by anyone other than main
func Close() error {
	if mariaCon != nil {
		if err := mariaCon.Close(); err != nil {
			return eris.Wrap(err, "Closing DB")
		}
	} else {
		slog.Info("MariaDB Connection is already closed or is not opened in the first place")
	}

	if loggerMariaCon != nil {
		if err := loggerMariaCon.Close(); err != nil {
			return eris.Wrap(err, "Closing Logger DB")
		}
	} else {
		slog.Info("Logger MariaDB Connection is already closed or is not opened in the first place")
	}

	return nil
}
