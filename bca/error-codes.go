package bca

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
	"4002400": "Bad Request",                           // Kesalahan dalam request atau parsing
	"5002400": "General Error",                         // Kesalahan umum di server
}
