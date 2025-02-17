package bank_integration_logger

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/rotisserie/eris"
	biModel "github.com/voxtmault/bank-integration/models"
	biStorage "github.com/voxtmault/bank-integration/storage"
	biUtil "github.com/voxtmault/bank-integration/utils"
)

var logChan chan *biModel.BankLogV2
var dbCon *sql.DB
var logVal *validator.Validate
var cancelFunc context.CancelFunc
var logMutex sync.Mutex

func InitLogger() {
	// init the channel
	logChan = make(chan *biModel.BankLogV2)
	dbCon = biStorage.GetLoggerDBConnection()
	logVal = biUtil.GetValidator()

	ctx, cancel := context.WithCancel(context.Background())
	cancelFunc = cancel

	// start the log worker
	go LogWorker(ctx)
}

func CloseLogger() {
	cancelFunc()
	close(logChan)
}

func LogRequest(log *biModel.BankLogV2) {
	logMutex.Lock()
	defer logMutex.Unlock()
	logChan <- log
}

func LogWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			slog.Debug("logger worker is stopped")
			return
		case log := <-logChan:
			slog.Info("received log", "log", log)
			if log == nil {
				slog.Warn("nil log received, skipping")
				return
			}

			if log.EndAt.IsZero() {
				// Meaning that http request were never sent, skipping
				slog.Debug("end time is nill / zero, no http request is sent. skipping...")
				continue
			}

			// Validate the obj before passing it to the core function
			if err := logVal.Struct(log); err != nil {
				slog.Error("failed to validate log", "reason", err)
			}
			slog.Debug("here")

			if log.ClientIP != "" {
				if err := logBankIngress(context.Background(), log); err != nil {
					slog.Error("failed to log bank ingress", "reason", err)
				}
			} else {
				if err := logBankEgress(context.Background(), log); err != nil {
					slog.Error("failed to log bank egress", "reason", err)
				}
			}
		}
	}
}

func GetIngressLogs(ctx context.Context, filter *biModel.BankLogSearchFilter) ([]*biModel.BankLogPublic, *biModel.PaginationMetadata, error) {
	arrObj := []*biModel.BankLogPublic{}
	pageMeta := biModel.PaginationMetadata{
		CurrentLimit: filter.Limit,
		CurrentPage:  filter.PageNumber,
	}

	var args []interface{}

	statement := `
	SELECT id, id_bank, id_feature, latency, response_code, client_ip, http_method, protocol, uri,
		   request_header, request_payload, response_header, response_payload, created_at
	FROM bank_ingress_logs
	WHERE 1 = 1
	`
	pagination := `
	SELECT COUNT(id) FROM bank_ingress_logs WHERE 1 = 1
	`

	if filter.ID > 0 {
		statement += " AND id = ?"
		pagination += " AND id = ?"
		args = append(args, filter.ID)
	}
	if filter.IDBank > 0 {
		statement += " AND id_bank = ?"
		pagination += " AND id_bank = ?"
		args = append(args, filter.IDBank)
	}
	if filter.IDFeature > 0 {
		statement += " AND id_feature = ?"
		pagination += " AND id_feature = ?"
		args = append(args, filter.IDFeature)
	}
	if filter.ClientIP != "" {
		statement += " AND client_ip = ?"
		pagination += " AND client_ip = ?"
		args = append(args, filter.ClientIP)
	}
	if filter.HTTPMethod != "" {
		statement += " AND http_method = ?"
		pagination += " AND http_method = ?"
		args = append(args, filter.HTTPMethod)
	}
	if filter.ResponseCode > 0 {
		statement += " AND response_code = ?"
		pagination += " AND response_code = ?"
		args = append(args, filter.ResponseCode)
	}
	if filter.URI != "" {
		statement += " AND uri = ?"
		pagination += " AND uri = ?"
		args = append(args, filter.URI)
	}
	if filter.StartDateRange != "" {
		statement += " AND DATE(created_at) >= ?"
		pagination += " AND DATE(created_at) >= ?"
		args = append(args, filter.StartDateRange)
	}
	if filter.EndDateRange != "" {
		statement += " AND DATE(created_at) <= ?"
		pagination += " AND DATE(created_at) <= ?"
		args = append(args, filter.EndDateRange)
	}

	offset := (filter.PageNumber - 1) * filter.Limit
	statement += " ORDER BY created_at DESC LIMIT ? OFFSET ? "

	args = append(args, filter.Limit, offset)
	rows, err := dbCon.QueryContext(ctx, statement, args...)
	if err != nil {
		slog.Error("failed to query context", "reason", err)
		return nil, nil, eris.Wrap(err, "failed to query context")
	}
	defer rows.Close()

	for rows.Next() {
		var obj biModel.BankLogPublic
		var reqHead, reqBod, resHead, resBod string
		if err = rows.Scan(
			&obj.ID, &obj.Bank.ID, &obj.RelatedFeature.ID, &obj.Latency, &obj.ResponseCode, &obj.ClientIP,
			&obj.HTTPMethod, &obj.Protocol, &obj.URI, &reqHead, &reqBod, &resHead, &resBod, &obj.CreatedAt,
		); err != nil {
			slog.Error("failed to scan rows", "reason", err)
			return nil, nil, eris.Wrap(err, "failed to scan rows")
		}

		// Unmarshall the JSON strings to map[string]interface{}

		if err = json.Unmarshal([]byte(reqHead), &obj.RequestHeader); err != nil {
			slog.Error("failed to unmarshal request header", "reason", err)
			return nil, nil, eris.Wrap(err, "failed to unmarshal request header")
		}
		if err = json.Unmarshal([]byte(reqBod), &obj.RequestBody); err != nil {
			slog.Error("failed to unmarshal request body", "reason", err)
			return nil, nil, eris.Wrap(err, "failed to unmarshal request body")
		}
		if err = json.Unmarshal([]byte(resHead), &obj.ResponseHeader); err != nil {
			slog.Error("failed to unmarshal response header", "reason", err)
			return nil, nil, eris.Wrap(err, "failed to unmarshal response header")
		}
		if err = json.Unmarshal([]byte(resBod), &obj.ResponsePayload); err != nil {
			slog.Error("failed to unmarshal response payload", "reason", err)
			return nil, nil, eris.Wrap(err, "failed to unmarshal response payload")
		}

		obj.GetHelperName()
		arrObj = append(arrObj, &obj)
	}

	if err := ProcessPaginationRequest(ctx, dbCon, pagination, args, &pageMeta); err != nil {
		slog.Error("failed to process pagination request", "reason", err)
		return nil, nil, eris.Wrap(err, "failed to process pagination request")
	}

	return arrObj, &pageMeta, nil
}

func GetEgressLogs(ctx context.Context, filter *biModel.BankLogSearchFilter) ([]*biModel.BankLogPublic, *biModel.PaginationMetadata, error) {
	arrObj := []*biModel.BankLogPublic{}
	pageMeta := biModel.PaginationMetadata{
		CurrentLimit: filter.Limit,
		CurrentPage:  filter.PageNumber,
	}

	var args []interface{}

	statement := `
	SELECT id, id_bank, id_feature, latency, response_code, host_ip, http_method, protocol, uri,
		   request_header, request_payload, response_header, response_payload, created_at
	FROM bank_egress_logs
	WHERE 1 = 1
	`
	pagination := `
	SELECT COUNT(id) FROM bank_egress_logs WHERE 1 = 1
	`

	if filter.ID > 0 {
		statement += " AND id = ?"
		pagination += " AND id = ?"
		args = append(args, filter.ID)
	}
	if filter.IDBank > 0 {
		statement += " AND id_bank = ?"
		pagination += " AND id_bank = ?"
		args = append(args, filter.IDBank)
	}
	if filter.IDFeature > 0 {
		statement += " AND id_feature = ?"
		pagination += " AND id_feature = ?"
		args = append(args, filter.IDFeature)
	}
	if filter.HostIP != "" {
		statement += " AND host_ip = ?"
		pagination += " AND host_ip = ?"
		args = append(args, filter.HostIP)
	}
	if filter.HTTPMethod != "" {
		statement += " AND http_method = ?"
		pagination += " AND http_method = ?"
		args = append(args, filter.HTTPMethod)
	}
	if filter.ResponseCode > 0 {
		statement += " AND response_code = ?"
		pagination += " AND response_code = ?"
		args = append(args, filter.ResponseCode)
	}
	if filter.URI != "" {
		statement += " AND uri = ?"
		pagination += " AND uri = ?"
		args = append(args, filter.URI)
	}
	if filter.StartDateRange != "" {
		statement += " AND DATE(created_at) >= ?"
		pagination += " AND DATE(created_at) >= ?"
		args = append(args, filter.StartDateRange)
	}
	if filter.EndDateRange != "" {
		statement += " AND DATE(created_at) <= ?"
		pagination += " AND DATE(created_at) <= ?"
		args = append(args, filter.EndDateRange)
	}

	offset := (filter.PageNumber - 1) * filter.Limit
	statement += " ORDER BY created_at DESC LIMIT ? OFFSET ? "

	args = append(args, filter.Limit, offset)
	rows, err := dbCon.QueryContext(ctx, statement, args...)
	if err != nil {
		slog.Error("failed to query context", "reason", err)
		return nil, nil, eris.Wrap(err, "failed to query context")
	}
	defer rows.Close()

	for rows.Next() {
		var obj biModel.BankLogPublic
		var reqHead, reqBod, resHead, resBod string
		if err = rows.Scan(
			&obj.ID, &obj.Bank.ID, &obj.RelatedFeature.ID, &obj.Latency, &obj.ResponseCode, &obj.HostIP,
			&obj.HTTPMethod, &obj.Protocol, &obj.URI, &reqHead, &reqBod, &resHead, &resBod, &obj.CreatedAt,
		); err != nil {
			slog.Error("failed to scan rows", "reason", err)
			return nil, nil, eris.Wrap(err, "failed to scan rows")
		}

		// Unmarshall the JSON strings to map[string]interface{}

		if err = json.Unmarshal([]byte(reqHead), &obj.RequestHeader); err != nil {
			slog.Error("failed to unmarshal request header", "reason", err)
			return nil, nil, eris.Wrap(err, "failed to unmarshal request header")
		}
		if err = json.Unmarshal([]byte(reqBod), &obj.RequestBody); err != nil {
			slog.Error("failed to unmarshal request body", "reason", err)
			return nil, nil, eris.Wrap(err, "failed to unmarshal request body")
		}
		if err = json.Unmarshal([]byte(resHead), &obj.ResponseHeader); err != nil {
			slog.Error("failed to unmarshal response header", "reason", err)
			return nil, nil, eris.Wrap(err, "failed to unmarshal response header")
		}
		if err = json.Unmarshal([]byte(resBod), &obj.ResponsePayload); err != nil {
			slog.Error("failed to unmarshal response payload", "reason", err)
			return nil, nil, eris.Wrap(err, "failed to unmarshal response payload")
		}

		obj.GetHelperName()
		arrObj = append(arrObj, &obj)
	}

	if err := ProcessPaginationRequest(ctx, dbCon, pagination, args, &pageMeta); err != nil {
		slog.Error("failed to process pagination request", "reason", err)
		return nil, nil, eris.Wrap(err, "failed to process pagination request")
	}

	return arrObj, &pageMeta, nil
}

// Internal Functions

func logBankIngress(ctx context.Context, log *biModel.BankLogV2) error {
	slog.Debug("ingress log")
	tx, err := dbCon.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("failed to begin transaction", "reason", err)
		return eris.Wrap(err, "failed to begin transaction")
	}

	log.Latency = log.EndAt.Sub(log.BeginAt).String()

	statement := `
	INSERT INTO bank_ingress_logs (id_bank, id_feature, latency, response_code, client_ip, http_method,
								   protocol, uri, request_header, request_payload, response_header,
								   response_payload)
	VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
	`
	if _, err = tx.ExecContext(ctx, statement,
		log.IDBank, log.IDFeature, log.Latency, log.ResponseCode, log.ClientIP, log.HTTPMethod, log.Protocol,
		log.URI, log.RequestHeader, log.RequestBody, log.ResponseHeader, log.ResponseBody,
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

	slog.Debug("done ingress")
	return nil
}

func logBankEgress(ctx context.Context, log *biModel.BankLogV2) error {
	slog.Debug("egress log")
	tx, err := dbCon.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("failed to begin transaction", "reason", err)
		return eris.Wrap(err, "failed to begin transaction")
	}

	log.Latency = log.EndAt.Sub(log.BeginAt).String()

	statement := `
	INSERT INTO bank_egress_logs (id_bank, id_feature, latency, response_code, host_ip, http_method,
								  protocol, uri, request_header, request_payload, response_header,
								  response_payload)
	VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
	`
	if _, err = tx.ExecContext(ctx, statement,
		log.IDBank, log.IDFeature, log.Latency, log.ResponseCode, log.HostIP, log.HTTPMethod, log.Protocol,
		log.URI, log.RequestHeader, log.RequestBody, log.ResponseHeader, log.ResponseBody,
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

func ProcessPaginationRequest(ctx context.Context, con *sql.DB, statement string, args []interface{}, metadata *biModel.PaginationMetadata) error {

	// Update the pagination metadata
	args = RemoveLastTwoItems(args)
	if err := con.QueryRowContext(ctx, statement, args...).Scan(&metadata.TotalRecords); err != nil {
		slog.Error("failed to get queue data count", "error", err)
		return eris.Wrap(err, "failed to get queue data count")
	}

	metadata.TotalPages = (metadata.TotalRecords + metadata.CurrentLimit - 1) / metadata.CurrentLimit

	return nil
}

func RemoveLastTwoItems[T any](slice []T) []T {
	if len(slice) < 2 {
		return []T{} // Return an empty slice if there are fewer than 2 items
	}
	return slice[:len(slice)-2]
}
