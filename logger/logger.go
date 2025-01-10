package bank_integration_logger

import (
	"context"
	"log/slog"

	"github.com/rotisserie/eris"
	biModel "github.com/voxtmault/bank-integration/models"
	biStorage "github.com/voxtmault/bank-integration/storage"
)

type contextKey string

const BankLogCtxKey contextKey = "bank_log"

func LogBankIngress(ctx context.Context, log *biModel.BankLog) error {
	con := biStorage.GetDBConnection()
	tx, err := con.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("failed to begin transaction", "reason", err)
		return eris.Wrap(err, "failed to begin transaction")
	}

	statement := `
	INSERT INTO bank_ingress (client_ip, latency, http_method, protocol, uri, response_code,
							  response_message, response_content, request_parameter, 
							  request_body)
	VALUES (?,?,?,?,?,?,?,?,?,?)
	`
	if _, err = tx.ExecContext(ctx, statement, log.ClientIP, log.Latency, log.HTTPMethod, log.Protocol,
		log.URI, log.ResponseCode, log.ResponseMessage, log.ResponseContent, log.RequestParameter,
		log.RequestBody,
	); err != nil {
		tx.Rollback()
		slog.Error("failed to exec statement", "reason", err)
		return eris.Wrap(err, "failed to exec statement")
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		slog.Error("failed to commit transaction", "reason", err)
		return eris.Wrap(err, "failed to commit transaction")
	}
	return nil
}

func LogBankEgress(ctx context.Context, log *biModel.BankLog) error {
	con := biStorage.GetDBConnection()
	tx, err := con.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("failed to begin transaction", "reason", err)
		return eris.Wrap(err, "failed to begin transaction")
	}

	statement := `
	INSERT INTO bank_egress (host_ip, latency, http_method, protocol, uri, response_code,
							 response_message, response_content, created_at, request_parameter, 
							 request_body)
	VALUES (?,?,?,?,?,?,?,?,?,?,?)
	`
	if _, err = tx.ExecContext(ctx, statement, log.HostIP, log.Latency, log.HTTPMethod, log.Protocol,
		log.URI, log.ResponseCode, log.ResponseMessage, log.ResponseContent, log.CreatedAt,
		log.RequestParameter, log.RequestBody,
	); err != nil {
		tx.Rollback()
		slog.Error("failed to exec statement", "reason", err)
		return eris.Wrap(err, "failed to exec statement")
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		slog.Error("failed to commit transaction", "reason", err)
		return eris.Wrap(err, "failed to commit transaction")
	}
	return nil
}
