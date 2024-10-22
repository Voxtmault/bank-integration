package bank_integration

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/rotisserie/eris"
	biStorage "github.com/voxtmault/bank-integration/storage"
	biUtil "github.com/voxtmault/bank-integration/utils"
)

// LoadAuthenticatedBanks will first retrieve the registered banks client credentials from a DB
// and then load them up into redis for faster lookup
func LoadAuthenticatedBanks(db *sql.DB, rdb *biStorage.RedisInstance) error {

	statement := `
	SELECT client_id, client_secret
	FROM authenticated_banks
	WHERE deleted_at IS NULL
	`
	rows, err := db.Query(statement)
	if err != nil {
		slog.Debug("Error querying authenticated banks")
		return eris.Wrap(err, "querying authenticated banks")
	}
	defer rows.Close()

	var clientId, clientSecret string
	for rows.Next() {
		if err := rows.Scan(&clientId, &clientSecret); err != nil {
			return eris.Wrap(err, "scanning rows")
		}

		if err := rdb.RDB.HSet(context.Background(), biUtil.ClientCredentialsRedis, clientId, clientSecret).Err(); err != nil {
			return eris.Wrap(err, "saving client credentials to redis")
		}
	}

	return nil
}
