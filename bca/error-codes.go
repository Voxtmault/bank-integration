package bca

import (
	"net/http"

	"github.com/voxtmault/bank-integration/models"
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
	BCACommonResponseMessageUnauthorizedSignature                  = BCACommonResponseMessage("Unauthorized [Signature]")
	BCACommonResponseMessageUnauthorizedStringToSign               = BCACommonResponseMessage("Unauthorized [Signature]")
	BCACommonResponseMessageUnauthorizedUnknownClient              = BCACommonResponseMessage("Unauthorized [Unknown client]")
	BCACommonResponseMessageMissingMandatoryField                  = BCACommonResponseMessage("Invalid Mandatory Field")
	BCACommonResponseMessageInvalidFieldFormat                     = BCACommonResponseMessage("Invalid Field Format")
	BCACommonResponseMessageDuplicateExternalID                    = BCACommonResponseMessage("Conflict")
	BCACommonResponseMessageVAPaid                                 = BCACommonResponseMessage("Paid Bill")
	BCACommonResponseMessageVAExpired                              = BCACommonResponseMessage("Invalid Bill/Virtual Account")
	BCACommonResponseMessageVANotFound                             = BCACommonResponseMessage("Invalid Bill/Virtual Account [Not Found]")
	BCACommonResponseMessageRequestParseError                      = BCACommonResponseMessage("Bad Request")
	BCACommonResponseMessageResponseParseError                     = BCACommonResponseMessage("Bad Request")
	BCACommonResponseMessageDuplicateExternalIDAndPaymentRequestID = BCACommonResponseMessage("Inconsistent Request")
	BCACommonResponseMessageGeneralError                           = BCACommonResponseMessage("General Error")
)

// Payment Flag Expected Partner Responses
var (
	BCAPaymentFlagResponseSuccess = models.BCAResponse{
		HTTPStatusCode:  http.StatusOK,
		ResponseCode:    "2002500",
		ResponseMessage: BCACommonResponseMessageSuccess.ToString(),
	}
	BCAPaymentFlagInvalidToken = models.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4012501",
		ResponseMessage: BCACommonResponseMessageInvalidToken.ToString(),
	}
	BCAPaymentFlagResponseUnauthorizedSignature = models.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4012500",
		ResponseMessage: BCACommonResponseMessageUnauthorizedSignature.ToString(),
	}
	BCAPaymentFlagResponseUnauthorizedStringToSign = models.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4012500",
		ResponseMessage: BCACommonResponseMessageUnauthorizedStringToSign.ToString(),
	}
	BCAPaymentFlagResponseUnauthorizedUnknownClient = models.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4012500",
		ResponseMessage: BCACommonResponseMessageUnauthorizedUnknownClient.ToString(),
	}
	BCAPaymentFlagResponseMissingMandatoryField = models.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4002502",
		ResponseMessage: BCACommonResponseMessageMissingMandatoryField.ToString(),
	}
	BCAPaymentFlagResponseInvalidFieldFormat = models.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4002501",
		ResponseMessage: BCACommonResponseMessageInvalidFieldFormat.ToString(),
	}
	BCAPaymentFlagResponseDuplicateExternalID = models.BCAResponse{
		HTTPStatusCode:  http.StatusConflict,
		ResponseCode:    "4092500",
		ResponseMessage: BCACommonResponseMessageDuplicateExternalID.ToString(),
	}
	BCAPaymentFlagResponseVAPaid = models.BCAResponse{
		HTTPStatusCode:  http.StatusNotFound,
		ResponseCode:    "4042514",
		ResponseMessage: BCACommonResponseMessageVAPaid.ToString(),
	}
	BCAPaymentFlagResponseDuplicateExternalIDAndPaymentRequestID = models.BCAResponse{
		HTTPStatusCode:  http.StatusConflict,
		ResponseCode:    "4042518",
		ResponseMessage: BCACommonResponseMessageDuplicateExternalIDAndPaymentRequestID.ToString(),
	}
	BCAPaymentFlagResponseVAExpired = models.BCAResponse{
		HTTPStatusCode:  http.StatusNotFound,
		ResponseCode:    "4042519",
		ResponseMessage: BCACommonResponseMessageVAExpired.ToString(),
	}
	BCAPaymentFlagResponseVANotFound = models.BCAResponse{
		HTTPStatusCode:  http.StatusNotFound,
		ResponseCode:    "4042512",
		ResponseMessage: BCACommonResponseMessageVANotFound.ToString(),
	}
	BCAPaymentFlagResponseRequestParseError = models.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4002500",
		ResponseMessage: BCACommonResponseMessageRequestParseError.ToString(),
	}
	BCAPaymentFlagResponseResponseParseError = models.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4002500",
		ResponseMessage: BCACommonResponseMessageRequestParseError.ToString(),
	}
	BCAPaymentFlagResponseGeneralError = models.BCAResponse{
		HTTPStatusCode:  http.StatusInternalServerError,
		ResponseCode:    "5002500",
		ResponseMessage: BCACommonResponseMessageGeneralError.ToString(),
	}
)

// Bill Inquiry Expected Partner Responses
var (
	BCABillInquiryResponseSuccess = models.BCAResponse{
		HTTPStatusCode:  http.StatusOK,
		ResponseCode:    "2002400",
		ResponseMessage: BCACommonResponseMessageSuccess.ToString(),
	}
	BCABillInquiryInvalidToken = models.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4012401",
		ResponseMessage: BCACommonResponseMessageInvalidToken.ToString(),
	}
	BCABillInquiryResponseUnauthorizedSignature = models.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4012400",
		ResponseMessage: BCACommonResponseMessageUnauthorizedSignature.ToString(),
	}
	BCABillInquiryResponseUnauthorizedStringToSign = models.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4012400",
		ResponseMessage: BCACommonResponseMessageUnauthorizedStringToSign.ToString(),
	}
	BCABillInquiryResponseUnauthorizedUnknownClient = models.BCAResponse{
		HTTPStatusCode:  http.StatusUnauthorized,
		ResponseCode:    "4012400",
		ResponseMessage: BCACommonResponseMessageUnauthorizedUnknownClient.ToString(),
	}
	BCABillInquiryResponseMissingMandatoryField = models.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4002402",
		ResponseMessage: BCACommonResponseMessageMissingMandatoryField.ToString(),
	}
	BCABillInquiryResponseInvalidFieldFormat = models.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4002401",
		ResponseMessage: BCACommonResponseMessageInvalidFieldFormat.ToString(),
	}
	BCABillInquiryResponseDuplicateExternalID = models.BCAResponse{
		HTTPStatusCode:  http.StatusConflict,
		ResponseCode:    "4092400",
		ResponseMessage: BCACommonResponseMessageDuplicateExternalID.ToString(),
	}
	BCABillInquiryResponseVAPaid = models.BCAResponse{
		HTTPStatusCode:  http.StatusNotFound,
		ResponseCode:    "4042414",
		ResponseMessage: BCACommonResponseMessageVAPaid.ToString(),
	}
	BCABillInquiryResponseVAExpired = models.BCAResponse{
		HTTPStatusCode:  http.StatusNotFound,
		ResponseCode:    "4042419",
		ResponseMessage: BCACommonResponseMessageVAExpired.ToString(),
	}
	BCABillInquiryResponseVANotFound = models.BCAResponse{
		HTTPStatusCode:  http.StatusNotFound,
		ResponseCode:    "4042412",
		ResponseMessage: BCACommonResponseMessageVANotFound.ToString(),
	}
	BCABillInquiryResponseRequestParseError = models.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4002400",
		ResponseMessage: BCACommonResponseMessageRequestParseError.ToString(),
	}
	BCABillInquiryResponseResponseParseError = models.BCAResponse{
		HTTPStatusCode:  http.StatusBadRequest,
		ResponseCode:    "4002400",
		ResponseMessage: BCACommonResponseMessageRequestParseError.ToString(),
	}
	BCABillInquiryResponseGeneralError = models.BCAResponse{
		HTTPStatusCode:  http.StatusInternalServerError,
		ResponseCode:    "5002400",
		ResponseMessage: BCACommonResponseMessageGeneralError.ToString(),
	}
)
