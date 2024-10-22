package bank_integration_utils

import (
	"log"
	"testing"
)

func TestGenerateClientCredential(t *testing.T) {
	clientCred := ClientCredential{}

	id, secret := clientCred.GenerateClientCredential()
	log.Println("ID: ", id)
	log.Println("Secret: ", secret)
}
