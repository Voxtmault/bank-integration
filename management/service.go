package bank_integration_management

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	biInterfaces "github.com/voxtmault/bank-integration/interfaces"
	biModel "github.com/voxtmault/bank-integration/models"
	biStorage "github.com/voxtmault/bank-integration/storage"
	biUtil "github.com/voxtmault/bank-integration/utils"
)

type BankIntegrationManagement struct {
	DB  *sql.DB
	RDB *biStorage.RedisInstance
}

var _ biInterfaces.Management = &BankIntegrationManagement{}

var service *BankIntegrationManagement

func NewBankIntegrationManagement(db *sql.DB, rdb *biStorage.RedisInstance) (*BankIntegrationManagement, error) {
	service = &BankIntegrationManagement{
		DB:  db,
		RDB: rdb,
	}

	if err := service.StartUp(context.Background()); err != nil {
		slog.Error("startup", "reason", err)
		return nil, eris.Wrap(err, "startup")
	}

	return service, nil
}

func GetManagementService() biInterfaces.Management {
	return service
}

func (s *BankIntegrationManagement) GetPartneredBanks(ctx context.Context) ([]*biModel.PartneredBank, error) {
	var arrObj []*biModel.PartneredBank

	statement := `
	SELECT id, bank_name, default_picture_path, partnership_status, created_at, COALESCE(updated_at, '')
	FROM partnered_banks
	WHERE deleted_at IS NULL
	`
	rows, err := s.DB.QueryContext(ctx, statement)
	if err != nil {
		return nil, eris.Wrap(err, "querying authenticated banks")
	}
	defer rows.Close()

	for rows.Next() {
		obj := biModel.PartneredBank{}.Default()
		if err = rows.Scan(
			&obj.ID, &obj.BankName, &obj.DefaultPicturePath, &obj.PartnershipStatus, &obj.CreatedAt, &obj.UpdatedAt,
		); err != nil {
			return nil, eris.Wrap(err, "scanning rows")
		}

		obj.IntegratedFeature, err = s.GetBankIntegratedFeatures(ctx, obj.ID)
		if err != nil {
			return nil, eris.Wrap(err, "getting bank integrated features")
		}

		slog.Debug("integrated features", "features", obj.IntegratedFeature)

		obj.PaymentMethod, err = s.GetBankPaymentMethods(ctx, obj.ID)
		if err != nil {
			return nil, eris.Wrap(err, "getting bank payment methods")
		}

		obj.ClientCredentials, err = s.GetBankClientCredentials(ctx, obj.ID)
		if err != nil {
			return nil, eris.Wrap(err, "getting bank client credentials")
		}

		arrObj = append(arrObj, &obj)
	}

	return arrObj, nil
}

func (s *BankIntegrationManagement) RegisterPartneredBank(ctx context.Context, obj *biModel.PartneredBankAdd) (*biModel.PartneredBank, error) {
	resObj := biModel.PartneredBank{}.Default()
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		tx.Rollback()
		return nil, eris.Wrap(err, "begin transaction")
	}

	resObj.BankName = obj.BankName
	resObj.DefaultPicturePath = obj.DefaultPicturePath
	resObj.PartnershipStatus = true
	resObj.CreatedAt = time.Now().Format(time.DateTime)

	statement := `
	INSERT INTO partnered_banks (bank_name, default_picture_path, partnership_status, created_at)
	VALUES(?,?,?,?)
	`
	result, err := tx.ExecContext(ctx, statement, obj.BankName, obj.DefaultPicturePath, true, resObj.CreatedAt)
	if err != nil {
		tx.Rollback()
		return nil, eris.Wrap(err, "executing statement")
	}

	id, _ := result.LastInsertId()
	resObj.ID = uint(id)

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return nil, eris.Wrap(err, "commit transaction")
	}

	// Register the client credentials

	creds, err := s.AddBankClientCredential(ctx, resObj.ID, "-")
	if err != nil {
		s.compensateDeleteRegisteredBanks(ctx, resObj.ID)
		return nil, eris.Wrap(err, "adding client credentials")
	}

	resObj.ClientCredentials = append(resObj.ClientCredentials, creds)

	return &resObj, nil
}

func (s *BankIntegrationManagement) UpdatePartneredBanks(ctx context.Context, obj *biModel.PartneredBank) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("begin transaction", "reason", err)
		return eris.Wrap(err, "begin transaction")
	}

	statement := `
	UPDATE partnered_banks SET bank_name = ?, default_picture_path = ?, partnership_status = ?
	WHERE id = ?
	`

	if _, err = tx.ExecContext(ctx, statement, obj.BankName, obj.DefaultPicturePath, obj.PartnershipStatus, obj.ID); err != nil {
		tx.Rollback()
		slog.Error("executing statement", "reason", err)
		return eris.Wrap(err, "executing statement")
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		slog.Error("commit transaction", "reason", err)
		return eris.Wrap(err, "commit transaction")
	}

	return nil
}

func (s *BankIntegrationManagement) DeletePartneredBank(ctx context.Context, idBank uint) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("begin transaction", "reason", err)
		return eris.Wrap(err, "begin transaction")
	}

	statement := `
	UPDATE partnered_banks SET deleted_at = CURRENT_TIMESTAMP() WHERE id = ?
	`
	if _, err = tx.ExecContext(ctx, statement, idBank); err != nil {
		tx.Rollback()
		slog.Error("executing statement", "reason", err)
		return eris.Wrap(err, "executing statement")
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		slog.Error("commit transaction", "reason", err)
		return eris.Wrap(err, "commit transaction")
	}

	return nil
}

func (s *BankIntegrationManagement) compensateDeleteRegisteredBanks(ctx context.Context, idBank uint) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("begin transaction", "reason", err)
		return eris.Wrap(err, "begin transaction")
	}

	statement := `
	DELETE FROM partnered_banks WHERE id = ?
	`
	if _, err = tx.ExecContext(ctx, statement, idBank); err != nil {
		tx.Rollback()
		slog.Error("executing statement", "reason", err)
		return eris.Wrap(err, "executing statement")
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		slog.Error("commit transaction", "reason", err)
		return eris.Wrap(err, "commit transaction")
	}

	return nil
}

// Bank Features

func (s *BankIntegrationManagement) GetBankIntegratedFeatures(ctx context.Context, idBank uint) ([]*biModel.IntegratedFeature, error) {
	arr := []*biModel.IntegratedFeature{}
	statement := `
	SELECT id, id_bank, id_feature, id_feature_type, feature_note, created_at, COALESCE(updated_at, '')
	FROM bank_integrated_features
	WHERE id_bank = ? AND deleted_at IS NULL
	`

	rows, err := s.DB.QueryContext(ctx, statement, idBank)
	if err != nil {
		slog.Error("querying bank integrated features", "reason", err)
		return nil, eris.Wrap(err, "querying bank integrated features")
	}
	defer rows.Close()

	for rows.Next() {
		obj := biModel.IntegratedFeature{}.Default()
		if err = rows.Scan(
			&obj.ID, &obj.Bank.ID, &obj.Feature.ID, &obj.FeatureType.ID, &obj.Note, &obj.CreatedAt,
			&obj.UpdatedAt,
		); err != nil {
			slog.Error("scanning rows", "reason", err)
			return nil, eris.Wrap(err, "scanning rows")
		}

		obj.GetHelperName()
		arr = append(arr, &obj)
	}

	return arr, nil
}

func (s *BankIntegrationManagement) EditBankIntegratedFeatures(ctx context.Context, arr []*biModel.IntegratedFeatureAdd) error {
	if len(arr) == 0 {
		slog.Warn("no features to edit")
		return nil
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("begin transaction", "reason", err)
		return eris.Wrap(err, "begin transaction")
	}

	existing := make(map[uint]*biModel.IntegratedFeature)
	statement := `
	SELECT id, id_feature, id_feature_type, feature_note
	FROM bank_integrated_features
	WHERE id_bank = ? AND deleted_at IS NULL
	`
	res, err := tx.QueryContext(ctx, statement, arr[0].IDBank)
	if err != nil {
		tx.Rollback()
		slog.Error("querying bank integrated features", "reason", err)
		return eris.Wrap(err, "querying bank integrated features")
	}
	defer res.Close()

	for res.Next() {
		obj := biModel.IntegratedFeature{}.Default()
		if err = res.Scan(&obj.ID, &obj.Feature.ID, &obj.FeatureType.ID, &obj.Note); err != nil {
			tx.Rollback()
			slog.Error("scanning rows", "reason", err)
			return eris.Wrap(err, "scanning rows")
		}

		existing[obj.Feature.ID] = &obj
	}

	statement = `
	INSERT INTO bank_integrated_features (id_bank, id_feature, id_feature_type, feature_note)
	VALUES (?,?,?,?)
	`
	insertStmt, err := tx.PrepareContext(ctx, statement)
	if err != nil {
		tx.Rollback()
		slog.Error("preparing statement", "reason", err)
		return eris.Wrap(err, "preparing statement")
	}
	defer insertStmt.Close()

	statement = `
	UPDATE bank_integrated_features SET id_feature_type = ?, feature_note = ?
	WHERE id = ?
	`
	updateStmt, err := tx.PrepareContext(ctx, statement)
	if err != nil {
		tx.Rollback()
		slog.Error("preparing statement", "reason", err)
		return eris.Wrap(err, "preparing statement")
	}
	defer updateStmt.Close()

	for _, new := range arr {
		if _, ok := existing[new.IDFeature]; ok {
			if _, err = updateStmt.ExecContext(ctx, new.IDFeatureType, new.Note, existing[new.IDFeature].ID); err != nil {
				tx.Rollback()
				slog.Error("executing update statement", "reason", err)
				return eris.Wrap(err, "executing statement")
			}
		} else {
			if _, err = insertStmt.ExecContext(ctx, new.IDBank, new.IDFeature, new.IDFeatureType, new.Note); err != nil {
				tx.Rollback()
				slog.Error("executing insert statement", "reason", err)
				return eris.Wrap(err, "executing statement")
			}
		}
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		slog.Error("commit transaction", "reason", err)
		return eris.Wrap(err, "commit transaction")
	}

	return nil
}

func (s *BankIntegrationManagement) DeleteBankIntegratedFeature(ctx context.Context, idBank, idFeature uint) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("begin transaction", "reason", err)
		return eris.Wrap(err, "begin transaction")
	}

	statement := `
	UPDATE bank_integrated_features SET deleted_at = CURRENT_TIMESTAMP() WHERE id = ? AND id_bank = ?
	`

	if _, err = tx.ExecContext(ctx, statement, idFeature, idBank); err != nil {
		tx.Rollback()
		slog.Error("executing statement", "reason", err)
		return eris.Wrap(err, "executing statement")
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		slog.Error("commit transaction", "reason", err)
		return eris.Wrap(err, "commit transaction")
	}

	return nil
}

// Bank Payment Methods

func (s *BankIntegrationManagement) GetBankPaymentMethods(ctx context.Context, idBank uint) ([]*biModel.PaymentMethod, error) {
	arr := []*biModel.PaymentMethod{}
	statement := `
	SELECT pm.id, pm.id_bank, pm.method_picture_path, pb.default_picture_path, pm.method_status, pm.created_at, COALESCE(pm.updated_at, '')
	FROM bank_payment_methods pm
	LEFT JOIN partnered_banks pb on pm.id_bank = pb.id
	WHERE pm.id_bank = ? AND pm.deleted_at IS NULL AND pb.deleted_at IS NULL
	`

	rows, err := s.DB.QueryContext(ctx, statement, idBank)
	if err != nil {
		slog.Error("querying bank payment method", "reason", err)
		return nil, eris.Wrap(err, "querying bank payment method")
	}
	defer rows.Close()

	var temp string
	for rows.Next() {
		obj := biModel.PaymentMethod{}.Default()
		if err = rows.Scan(
			&obj.ID, &obj.Bank.ID, &obj.PaymentMethod.ID, &obj.DefaultPicturePath, &temp,
			&obj.PaymentMethodStatus, &obj.CreatedAt, &obj.UpdatedAt,
		); err != nil {
			slog.Error("scanning rows", "reason", err)
			return nil, eris.Wrap(err, "scanning rows")
		}

		if obj.DefaultPicturePath == "" {
			// If the default picture path is empty, use the bank's default picture path
			obj.DefaultPicturePath = temp
		}

		obj.GetHelperName()
		arr = append(arr, &obj)
	}

	return arr, nil
}

func (s *BankIntegrationManagement) EditBankPamentMethods(ctx context.Context, arr []*biModel.PaymentMethodAdd) error {
	if len(arr) == 0 {
		slog.Warn("no payment methods to edit")
		return nil
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("begin transaction", "reason", err)
		return eris.Wrap(err, "begin transaction")
	}

	existing := make(map[uint]*biModel.PaymentMethod)
	statement := `
	SELECT id, method_name, method_picture_path, method_status
	FROM bank_payment_methods
	WHERE id_bank = ? AND deleted_at IS NULL
	`
	res, err := tx.QueryContext(ctx, statement, arr[0].IDBank)
	if err != nil {
		tx.Rollback()
		slog.Error("querying bank payment methods", "reason", err)
		return eris.Wrap(err, "querying bank payment methods")
	}
	defer res.Close()

	for res.Next() {
		obj := biModel.PaymentMethod{}.Default()
		if err = res.Scan(&obj.ID, &obj.PaymentMethod.ID, &obj.DefaultPicturePath, &obj.PaymentMethodStatus); err != nil {
			tx.Rollback()
			slog.Error("scanning rows", "reason", err)
			return eris.Wrap(err, "scanning rows")
		}

		existing[obj.PaymentMethod.ID] = &obj
	}

	statement = `
	INSERT INTO bank_payment_methods (id_bank, id_payment_method, method_picture_path)
	VALUES (?,?,?)
	`
	insertStmt, err := tx.PrepareContext(ctx, statement)
	if err != nil {
		tx.Rollback()
		slog.Error("preparing statement", "reason", err)
		return eris.Wrap(err, "preparing statement")
	}
	defer insertStmt.Close()

	statement = `
	UPDATE bank_payment_methods SET method_picture_path = ?, method_status = ?
	WHERE id = ?
	`
	updateStmt, err := tx.PrepareContext(ctx, statement)
	if err != nil {
		tx.Rollback()
		slog.Error("preparing statement", "reason", err)
		return eris.Wrap(err, "preparing statement")
	}
	defer updateStmt.Close()

	for _, new := range arr {
		if _, ok := existing[new.IDPaymentMethod]; ok {
			if _, err = updateStmt.ExecContext(ctx, new.DefaultPicturePath, new.PaymentMethodStatus, existing[new.IDPaymentMethod].ID); err != nil {
				tx.Rollback()
				slog.Error("executing update statement", "reason", err)
				return eris.Wrap(err, "executing statement")
			}
		} else {
			if _, err = insertStmt.ExecContext(ctx, new.IDBank, new.IDPaymentMethod, new.DefaultPicturePath); err != nil {
				tx.Rollback()
				slog.Error("executing insert statement", "reason", err)
				return eris.Wrap(err, "executing statement")
			}
		}
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		slog.Error("commit transaction", "reason", err)
		return eris.Wrap(err, "commit transaction")
	}

	return nil
}

func (s *BankIntegrationManagement) DeleteBankPaymentMethod(ctx context.Context, idBank, idPM uint) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("begin transaction", "reason", err)
		return eris.Wrap(err, "begin transaction")
	}

	statement := `
	UPDATE bank_payment_methods SET deleted_at = CURRENT_TIMESTAMP() WHERE id = ? AND id_bank = ?
	`

	if _, err = tx.ExecContext(ctx, statement, idPM, idBank); err != nil {
		tx.Rollback()
		slog.Error("executing statement", "reason", err)
		return eris.Wrap(err, "executing statement")
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		slog.Error("commit transaction", "reason", err)
		return eris.Wrap(err, "commit transaction")
	}

	return nil
}

// Bank Client Credentials

func (s *BankIntegrationManagement) GetBankClientCredentials(ctx context.Context, idBank uint) ([]*biModel.BankClientCredential, error) {
	arr := []*biModel.BankClientCredential{}
	statement := `
	SELECT id, id_bank, client_id, client_secret, credential_status, credential_note, created_at, COALESCE(updated_at, '')
	FROM bank_client_credentials
	WHERE id_bank = ? AND deleted_at IS NULL
	`
	rows, err := s.DB.QueryContext(ctx, statement, idBank)
	if err != nil {
		slog.Error("querying bank client credentials", "reason", err)
		return nil, eris.Wrap(err, "querying bank client credentials")
	}
	defer rows.Close()

	for rows.Next() {
		obj := biModel.BankClientCredential{}.Default()
		if err = rows.Scan(
			&obj.ID, &obj.Bank.ID, &obj.ClientID, &obj.ClientSecret, &obj.CredentialStatus, &obj.CredentialNote,
			&obj.CreatedAt, &obj.UpdatedAt,
		); err != nil {
			slog.Error("scanning rows", "reason", err)
			return nil, eris.Wrap(err, "scanning rows")
		}

		obj.GetHelperName()
		arr = append(arr, &obj)
	}

	return arr, nil
}

func (s *BankIntegrationManagement) AddBankClientCredential(ctx context.Context, idBank uint, note string) (*biModel.BankClientCredential, error) {
	obj := biModel.BankClientCredential{}.Default()
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("begin transaction", "reason", err)
		return nil, eris.Wrap(err, "begin transaction")
	}

	obj.CreatedAt = time.Now().Format(time.DateTime)
	obj.ClientID = uuid.New().String()
	obj.ClientSecret = uuid.New().String()
	obj.CredentialStatus = true
	obj.CredentialNote = note

	statement := `
	INSERT INTO bank_client_credentials (id_bank, client_id, client_secret, credential_status, credential_note, created_at)
	VALUES (?,?,?,?,?,?)
	`
	result, err := tx.ExecContext(ctx, statement, idBank, obj.ClientID, obj.ClientSecret, true, note, obj.CreatedAt)
	if err != nil {
		tx.Rollback()
		slog.Error("executing statement", "reason", err)
		return nil, eris.Wrap(err, "executing statement")
	}

	id, _ := result.LastInsertId()
	obj.ID = uint(id)

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		slog.Error("commit transaction", "reason", err)
		return nil, eris.Wrap(err, "commit transaction")
	}

	obj.GetHelperName()

	return &obj, nil
}

func (s *BankIntegrationManagement) EditBankClientCredential(ctx context.Context, obj *biModel.BankClientCredential) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("begin transaction", "reason", err)
		return eris.Wrap(err, "begin transaction")
	}

	statement := `
	UPDATE bank_client_credentials SET credential_status = ?, credential_note = ?
	WHERE id = ?
	`
	if _, err = tx.ExecContext(ctx, statement, obj.CredentialStatus, obj.CredentialStatus, obj.ID); err != nil {
		tx.Rollback()
		slog.Error("executing statement", "reason", err)
		return eris.Wrap(err, "executing statement")
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		slog.Error("commit transaction", "reason", err)
		return eris.Wrap(err, "commit transaction")
	}

	return nil
}

func (s *BankIntegrationManagement) DeleteBankClientCredential(ctx context.Context, idBank, idCC uint) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("begin transaction", "reason", err)
		return eris.Wrap(err, "begin transaction")
	}

	statement := `
	UPDATE bank_client_credentials SET deleted_at = CURRENT_TIMESTAMP() WHERE id = ? AND id_bank = ?
	`
	if _, err = tx.ExecContext(ctx, statement, idCC, idBank); err != nil {
		tx.Rollback()
		slog.Error("executing statement", "reason", err)
		return eris.Wrap(err, "executing statement")
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		slog.Error("commit transaction", "reason", err)
		return eris.Wrap(err, "commit transaction")
	}

	return nil
}

// Called On Startup
func (s *BankIntegrationManagement) StartUp(ctx context.Context) error {
	// List of items to do
	// 1. Load the Helper Table values to Redis
	// 2. Load the Client Credentials to Redis
	// 3. Load the Partnered Banks to Redis

	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	wg.Add(3)
	// 1.
	go func() {
		defer wg.Done()
		for _, item := range biUtil.StartupHelper {
			rows, err := s.DB.QueryContext(ctx, fmt.Sprintf(`SELECT id, name FROM %s WHERE deleted_at IS NULL`, item))
			if err != nil {
				errChan <- eris.Wrap(err, "querying helper table "+item)
				return
			}

			for rows.Next() {
				var obj biModel.Helper
				if err = rows.Scan(&obj.ID, &obj.Name); err != nil {
					errChan <- eris.Wrap(err, "scanning rows for "+item)
					return
				}

				// Store the values to Redis
				if err = s.RDB.RDB.HSet(ctx, item, fmt.Sprintf("%d", obj.ID), obj.Name).Err(); err != nil {
					errChan <- eris.Wrap(err, "storing to redis for "+item)
					return
				}
			}

			rows.Close()
		}

	}()

	// 2.
	go func() {
		defer wg.Done()
		statement := `
		SELECT client_id, client_secret, id
		FROM bank_client_credentials
		WHERE deleted_at IS NULL
		`
		rows, err := s.DB.QueryContext(ctx, statement)
		if err != nil {
			errChan <- eris.Wrap(err, "querying partnered banks")
			return
		}
		defer rows.Close()

		var cid, csec, bankId string
		for rows.Next() {
			if err = rows.Scan(&cid, &csec, &bankId); err != nil {
				errChan <- eris.Wrap(err, "scanning rows")
				return
			}

			// Store the values to Redis
			if err = s.RDB.RDB.HSet(ctx, biUtil.ClientCredentialsRedis, cid, csec).Err(); err != nil {
				errChan <- eris.Wrap(err, "storing to redis")
				return
			}

			// Store the values to Redis
			if err = s.RDB.RDB.HSet(ctx, biUtil.PartneredBanksCredentialsMapping, csec, bankId).Err(); err != nil {
				errChan <- eris.Wrap(err, "storing to redis")
				return
			}
		}
	}()

	// 3.
	go func() {
		defer wg.Done()
		statement := `
		SELECT id, bank_name FROM partnered_banks WHERE deleted_at IS NULL
		`
		rows, err := s.DB.QueryContext(ctx, statement)
		if err != nil {
			errChan <- eris.Wrap(err, "querying partnered banks")
			return
		}
		defer rows.Close()

		var obj biModel.Helper
		for rows.Next() {
			if err = rows.Scan(&obj.ID, &obj.Name); err != nil {
				errChan <- eris.Wrap(err, "scanning rows")
				return
			}

			// Store the values to Redis
			if err = s.RDB.RDB.HSet(ctx, biUtil.PartneredBanks, fmt.Sprintf("%d", obj.ID), obj.Name).Err(); err != nil {
				errChan <- eris.Wrap(err, "storing to redis")
				return
			}
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}
