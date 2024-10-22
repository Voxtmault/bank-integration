package bank_integration_management

import (
	"context"
	"database/sql"

	"github.com/rotisserie/eris"
	biInterfaces "github.com/voxtmault/bank-integration/interfaces"
	biModel "github.com/voxtmault/bank-integration/models"
	biStorage "github.com/voxtmault/bank-integration/storage"
	biUtil "github.com/voxtmault/bank-integration/utils"
)

type BankIntegrationManagement struct {
	DB  *sql.DB
	RDB *biStorage.RedisInstance
	GS  biUtil.ClientCredential
}

var _ biInterfaces.Management = &BankIntegrationManagement{}

func NewBankIntegrationManagement(db *sql.DB, rdb *biStorage.RedisInstance) *BankIntegrationManagement {
	return &BankIntegrationManagement{
		DB:  db,
		RDB: rdb,
		GS:  biUtil.ClientCredential{},
	}
}

func (s *BankIntegrationManagement) GetAuthenticatedBanks(ctx context.Context) ([]*biModel.AuthenticatedBank, error) {
	var arrObj []*biModel.AuthenticatedBank

	statement := `
	SELECT id, bank_name, note, created_at, updated_at
	FROM authenticated_banks
	WHERE deleted_at IS NULL
	`
	rows, err := s.DB.QueryContext(ctx, statement)
	if err != nil {
		return nil, eris.Wrap(err, "querying authenticated banks")
	}
	defer rows.Close()

	for rows.Next() {
		var obj biModel.AuthenticatedBank
		if err = rows.Scan(
			&obj.ID, &obj.BankName, &obj.PublicKeyPath, &obj.Note, &obj.CreatedAt, &obj.UpdatedAt,
		); err != nil {
			return nil, eris.Wrap(err, "scanning rows")
		}

		arrObj = append(arrObj, &obj)
	}

	return arrObj, nil
}

func (s *BankIntegrationManagement) RegisterBank(ctx context.Context, bankName string) (*biModel.BankClientCredential, error) {

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		tx.Rollback()
		return nil, eris.Wrap(err, "begin transaction")
	}

	id, secret := s.GS.GenerateClientCredential()

	statement := `
	INSERT INTO authenticated_banks (bank_name, client_id, client_secret)
	VALUES(?,?,?)
	`
	if _, err = tx.ExecContext(ctx, statement, bankName, id, secret); err != nil {
		tx.Rollback()
		return nil, eris.Wrap(err, "executing statement")
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return nil, eris.Wrap(err, "commit transaction")
	}

	return &biModel.BankClientCredential{
		ClientID:     id,
		ClientSecret: secret,
	}, nil
}

func (s *BankIntegrationManagement) UpdateRegisteredBank(ctx context.Context) error {
	return nil
}

func (s *BankIntegrationManagement) RevokeRegisteredBank(ctx context.Context) error {
	return nil
}
