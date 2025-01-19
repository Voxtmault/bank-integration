package bank_integration_internal

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/rotisserie/eris"
	biConfig "github.com/voxtmault/bank-integration/config"
	biInterfaces "github.com/voxtmault/bank-integration/interfaces"
	biModels "github.com/voxtmault/bank-integration/models"
	biStorage "github.com/voxtmault/bank-integration/storage"
	biUtil "github.com/voxtmault/bank-integration/utils"
)

type InternalService struct {
	// Configs
	Config *biConfig.BankingConfig

	// DB Connections
	DB  *sql.DB
	RDB *biStorage.RedisInstance
}

var _ biInterfaces.Internal = &InternalService{}

func NewInternalService(config *biConfig.BankingConfig, db *sql.DB, rdb *biStorage.RedisInstance) (*InternalService, error) {
	service := InternalService{
		Config: config,
		DB:     db,
		RDB:    rdb,
	}

	return &service, nil
}

func (i *InternalService) GetOrderVAInformation(ctx context.Context, idOrder uint) (*biModels.InternalVAInformation, error) {
	var err error
	var obj biModels.InternalVAInformation
	statement := `
	SELECT id_bank, virtualAccountNo, virtualAccountName, totalAmountValue, expired_date
	FROM va_request
	WHERE id_order = ? AND id_va_status = ?
	ORDER BY created_at DESC
	LIMIT 1
	`
	if err = i.DB.QueryRowContext(ctx, statement, idOrder, biUtil.VAStatusPending).Scan(
		&obj.IDBank,
		&obj.VANumber,
		&obj.VAAccountName,
		&obj.TotalAmount,
		&obj.ExpiredAt,
	); err != nil {
		return nil, eris.Wrap(err, "querying va request")
	}

	// Get the bank name from redis
	obj.BankName, err = i.RDB.RDB.HGet(ctx, biUtil.AuthenticatedBankNameRedis, strconv.Itoa(int(obj.IDBank))).Result()
	if err != nil {
		obj.BankName = ""
	}

	// Get the bank icon logo from redis
	obj.BankIconLink, err = i.RDB.RDB.HGet(ctx, biUtil.VendorsLogoRedis, strconv.Itoa(int(obj.IDBank))).Result()
	if err != nil {
		obj.BankIconLink = ""
	}

	return &obj, nil
}

func (i *InternalService) GetTopUpVAInformation(ctx context.Context, trxId uint) (*biModels.InternalVAInformation, error) {
	var err error
	var obj biModels.InternalVAInformation
	statement := `
	SELECT id_bank, virtualAccountNo, virtualAccountName, totalAmountValue, expired_date
	FROM va_request
	WHERE id_transaction = ? AND id_va_status = ?
	ORDER BY created_at DESC
	LIMIT 1
	`
	if err = i.DB.QueryRowContext(ctx, statement, trxId, biUtil.VAStatusPending).Scan(
		&obj.IDBank,
		&obj.VANumber,
		&obj.VAAccountName,
		&obj.TotalAmount,
		&obj.ExpiredAt,
	); err != nil {
		return nil, eris.Wrap(err, "querying va request")
	}

	// Get the bank name from redis
	obj.BankName, err = i.RDB.RDB.HGet(ctx, biUtil.AuthenticatedBankNameRedis, strconv.Itoa(int(obj.IDBank))).Result()
	if err != nil {
		obj.BankName = ""
	}

	// Get the bank icon logo from redis
	obj.BankIconLink, err = i.RDB.RDB.HGet(ctx, biUtil.VendorsLogoRedis, strconv.Itoa(int(obj.IDBank))).Result()
	if err != nil {
		obj.BankIconLink = ""
	}

	return &obj, nil
}
