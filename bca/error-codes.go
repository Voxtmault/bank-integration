package bca

import (
	"net/http"

	biModels "github.com/voxtmault/bank-integration/models"
)

var BcaErrorCodes = map[string]string{
	"4012401": "Invalid Token (B2B)",                   // Token tidak valid
	"4012400": "Unauthorized [Signature]",              // Unauthorized, Signature tidak sah
	"4012403": "Unauthorized [Unknown client]",         // Unauthorized, client tidak dikenal
	"4002402": "Invalid Mandatory Field",               // Mandatory field hilang
	"4002401": "Invalid Field Format",                  // Format field tidak valid
	"4092400": "Conflict",                              // Konflik, X-EXTERNAL-ID yang sama
	"2002400": "Success",                               // Request berhasil
	"4042414": "Paid Bill",                             // Tagihan sudah dibayar
	"4042419": "Invalid Bill/Virtual Account",          // Tagihan atau Virtual Account tidak valid/kedaluwarsa
	"4042412": "Invalid Bill/Virtual Account [Reason]", // Tagihan atau Virtual Account tidak ditemukan
	"4042512": "Invalid Bill/Virtual Account [Not Found]",
	"4002400": "Bad Request",   // Kesalahan dalam request atau parsing
	"5002400": "General Error", // Kesalahan umum di server
}

type BCACommonResponseMessage string

func (b BCACommonResponseMessage) ToString() string {
	return string(b)
}

// Common BCA Response Message Collections
var (
	BCACommonResponseMessageSuccess                                = BCACommonResponseMessage("Success")
	BCACommonResponseMessageInvalidToken                           = BCACommonResponseMessage("Invalid Token (B2B)")
	BCACommonResponseMessageUnauthorizedSignature                  = BCACommonResponseMessage("Unauthorized. [Signature]")
	BCACommonResponseMessageUnauthorizedStringToSign               = BCACommonResponseMessage("Unauthorized. [Signature]")
	BCACommonResponseMessageUnauthorizedUnknownClient              = BCACommonResponseMessage("Unauthorized. [Unknown client]")
	BCACommonResponseMessageUnauthorizedConnectionNotAllowed       = BCACommonResponseMessage("Unauthorized. [Connection not allowed]")
	BCACommonResponseMessageMissingMandatoryField                  = BCACommonResponseMessage("Invalid mandatory field")
	BCACommonResponseMessageInvalidFieldFormat                     = BCACommonResponseMessage("Invalid field format")
	BCACommonResponseMessageDuplicateExternalID                    = BCACommonResponseMessage("Conflict")
	BCACommonResponseMessageVAPaid                                 = BCACommonResponseMessage("Paid Bill")
	BCACommonResponseMessageInvalidAmount                          = BCACommonResponseMessage("Invalid Amount")
	BCACommonResponseMessageVAExpired                              = BCACommonResponseMessage("Invalid Bill/Virtual Account")
	BCACommonResponseMessageVANotFound                             = BCACommonResponseMessage("Invalid Bill/Virtual Account [Not Found]")
	BCACommonResponseMessageRequestParseError                      = BCACommonResponseMessage("Bad Request")
	BCACommonResponseMessageResponseParseError                     = BCACommonResponseMessage("Bad Request")
	BCACommonResponseMessageDuplicateExternalIDAndPaymentRequestID = BCACommonResponseMessage("Inconsistent Request")
	BCACommonResponseMessageGeneralError                           = BCACommonResponseMessage("General Error")
	BCACommonResponseMessageTimeout                                = BCACommonResponseMessage("Timeout")
)

// Payment Flag Expected Partner Responses
var (
	BCAPaymentFlagResponseSuccess = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusOK,
		ResponseCode:    "2002500",
		ResponseMessage: BCACommonResponseMessageSuccess.ToString(),
	}
	BCAPaymentFlagInvalidToken = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4012501",
		ResponseMessage: BCACommonResponseMessageInvalidToken.ToString(),
	}
	BCAPaymentFlagResponseUnauthorizedSignature = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4012500",
		ResponseMessage: BCACommonResponseMessageUnauthorizedSignature.ToString(),
	}
	BCAPaymentFlagResponseUnauthorizedStringToSign = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4012500",
		ResponseMessage: BCACommonResponseMessageUnauthorizedStringToSign.ToString(),
	}
	BCAPaymentFlagResponseUnauthorizedUnknownClient = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4012500",
		ResponseMessage: BCACommonResponseMessageUnauthorizedUnknownClient.ToString(),
	}
	BCAPaymentFlagResponseMissingMandatoryField = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4002502",
		ResponseMessage: BCACommonResponseMessageMissingMandatoryField.ToString(),
	}
	BCAPaymentFlagResponseInvalidFieldFormat = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4002501",
		ResponseMessage: BCACommonResponseMessageInvalidFieldFormat.ToString(),
	}
	BCAPaymentFlagResponseDuplicateExternalID = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusConflict,
		ResponseCode:    "4092500",
		ResponseMessage: BCACommonResponseMessageDuplicateExternalID.ToString(),
	}
	BCAPaymentFlagResponseVAPaid = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusNotFound,
		ResponseCode:    "4042514",
		ResponseMessage: BCACommonResponseMessageVAPaid.ToString(),
	}
	BCAPaymentFlagResponseDuplicateExternalIDAndPaymentRequestID = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusNotFound,
		ResponseCode:    "4042518",
		ResponseMessage: BCACommonResponseMessageDuplicateExternalIDAndPaymentRequestID.ToString(),
	}
	BCAPaymentFlagResponseVAExpired = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusNotFound,
		ResponseCode:    "4042519",
		ResponseMessage: BCACommonResponseMessageVAExpired.ToString(),
	}
	BCAPaymentFlagResponseVANotFound = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusNotFound,
		ResponseCode:    "4042512",
		ResponseMessage: BCACommonResponseMessageVANotFound.ToString(),
	}
	BCAPaymentFlagResponseInvalidAmount = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusNotFound,
		ResponseCode:    "4042513",
		ResponseMessage: BCACommonResponseMessageInvalidAmount.ToString(),
	}
	BCAPaymentFlagResponseRequestParseError = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4002500",
		ResponseMessage: BCACommonResponseMessageRequestParseError.ToString(),
	}
	BCAPaymentFlagResponseResponseParseError = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4002500",
		ResponseMessage: BCACommonResponseMessageRequestParseError.ToString(),
	}
	BCAPaymentFlagResponseGeneralError = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusInternalServerError,
		ResponseCode:    "5002500",
		ResponseMessage: BCACommonResponseMessageGeneralError.ToString(),
	}
)

// Bill Inquiry Expected Partner Responses
var (
	BCABillInquiryResponseSuccess = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusOK,
		ResponseCode:    "2002400",
		ResponseMessage: BCACommonResponseMessageSuccess.ToString(),
	}
	BCABillInquiryInvalidToken = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4012401",
		ResponseMessage: BCACommonResponseMessageInvalidToken.ToString(),
	}
	BCABillInquiryResponseUnauthorizedSignature = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4012400",
		ResponseMessage: BCACommonResponseMessageUnauthorizedSignature.ToString(),
	}
	BCABillInquiryResponseUnauthorizedStringToSign = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4012400",
		ResponseMessage: BCACommonResponseMessageUnauthorizedStringToSign.ToString(),
	}
	BCABillInquiryResponseUnauthorizedUnknownClient = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4012400",
		ResponseMessage: BCACommonResponseMessageUnauthorizedUnknownClient.ToString(),
	}
	BCABillInquiryResponseMissingMandatoryField = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4002402",
		ResponseMessage: BCACommonResponseMessageMissingMandatoryField.ToString(),
	}
	BCABillInquiryResponseInvalidFieldFormat = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4002401",
		ResponseMessage: BCACommonResponseMessageInvalidFieldFormat.ToString(),
	}
	BCABillInquiryResponseDuplicateExternalID = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusConflict,
		ResponseCode:    "4092400",
		ResponseMessage: BCACommonResponseMessageDuplicateExternalID.ToString(),
	}
	BCABillInquiryResponseVAPaid = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusNotFound,
		ResponseCode:    "4042414",
		ResponseMessage: BCACommonResponseMessageVAPaid.ToString(),
	}
	BCABillInquiryResponseVAExpired = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusNotFound,
		ResponseCode:    "4042419",
		ResponseMessage: BCACommonResponseMessageVAExpired.ToString(),
	}
	BCABillInquiryResponseVANotFound = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusNotFound,
		ResponseCode:    "4042412",
		ResponseMessage: BCACommonResponseMessageVANotFound.ToString(),
	}
	BCABillInquiryResponseRequestParseError = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4002400",
		ResponseMessage: BCACommonResponseMessageRequestParseError.ToString(),
	}
	BCABillInquiryResponseResponseParseError = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4002400",
		ResponseMessage: BCACommonResponseMessageRequestParseError.ToString(),
	}
	BCABillInquiryResponseGeneralError = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusInternalServerError,
		ResponseCode:    "5002400",
		ResponseMessage: BCACommonResponseMessageGeneralError.ToString(),
	}
)

// Auhentication Expected Partner Responses
var (
	BCAAuthResponseSuccess = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusOK,
		ResponseCode:    "2007300",
		ResponseMessage: "Successful",
	}
	BCAAuthInvalidFieldFormatClient = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4007301",
		ResponseMessage: BCACommonResponseMessageInvalidFieldFormat.ToString() + " [clientId/clientSecret/grantType]",
	}
	BCAAuthInvalidFieldFormatTimestamp = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4007301",
		ResponseMessage: BCACommonResponseMessageInvalidFieldFormat.ToString() + " [X-TIMESTAMP]",
	}
	BCAAUthInvalidMandatoryField = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4007302",
		ResponseMessage: BCACommonResponseMessageMissingMandatoryField.ToString(),
	}
	BCAAuthUnauthorizedSignature = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4017300",
		ResponseMessage: BCACommonResponseMessageUnauthorizedSignature.ToString(),
	}
	BCAAuthUnauthorizedUnknownClient = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4017300",
		ResponseMessage: BCACommonResponseMessageUnauthorizedUnknownClient.ToString(),
	}
	BCAAuthUnauthorizedConnectionNotAllowed = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4017300",
		ResponseMessage: BCACommonResponseMessageUnauthorizedConnectionNotAllowed.ToString(),
	}
	BCAAuthTimeout = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusRequestTimeout,
		ResponseCode:    "5047300",
		ResponseMessage: BCACommonResponseMessageTimeout.ToString(),
	}
	BCAAuthGeneralError = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusInternalServerError,
		ResponseCode:    "5007300",
		ResponseMessage: BCACommonResponseMessageGeneralError.ToString(),
	}
	BCAAuthConflict = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusConflict,
		ResponseCode:    "4097300",
		ResponseMessage: BCACommonResponseMessageDuplicateExternalID.ToString(),
	}
	BCAAuthInvalidToken = biModels.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4017301",
		ResponseMessage: BCACommonResponseMessageInvalidToken.ToString(),
	}
)
