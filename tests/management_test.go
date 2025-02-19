package bank_integration_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	biMgm "github.com/voxtmault/bank-integration/management"
	bank_integration_models "github.com/voxtmault/bank-integration/models"
	biStorage "github.com/voxtmault/bank-integration/storage"
)

func TestInitManagementService(t *testing.T) {
	if err := setup(); err != nil {
		t.Fatalf("error setting up test: %v", err)
	}
	_, err := biMgm.NewBankIntegrationManagement(biStorage.GetLoggerDBConnection(), biStorage.GetRedisInstance())
	if err != nil {
		t.Fatalf("error initializing management service: %v", err)
	}
}

func TestAddPartneredBanks(t *testing.T) {
	if err := setup(); err != nil {
		t.Fatalf("error setting up test: %v", err)
	}
	mgm, err := biMgm.NewBankIntegrationManagement(biStorage.GetLoggerDBConnection(), biStorage.GetRedisInstance())
	if err != nil {
		t.Fatalf("error initializing management service: %v", err)
	}

	newBank, err := mgm.RegisterPartneredBank(context.Background(), &bank_integration_models.PartneredBankAdd{
		BankName:           "Bank BCA",
		DefaultPicturePath: "/path/to/picture",
	})
	if err != nil {
		t.Fatalf("error adding partnered bank: %v", err)
	}

	slog.Debug("received bank info", "bank", newBank)
}

func TestGetPartneredBanks(t *testing.T) {
	if err := setup(); err != nil {
		t.Fatalf("error setting up test: %v", err)
	}
	mgm, err := biMgm.NewBankIntegrationManagement(biStorage.GetLoggerDBConnection(), biStorage.GetRedisInstance())
	if err != nil {
		t.Fatalf("error initializing management service: %v", err)
	}

	banks, err := mgm.GetPartneredBanks(context.Background())
	if err != nil {
		t.Fatalf("error getting partnered banks: %v", err)
	}

	for _, item := range banks {
		str, _ := json.Marshal(item)
		slog.Debug("received bank info", "bank", str)
	}
}

func TestEditPartneredBanks(t *testing.T) {
	if err := setup(); err != nil {
		t.Fatalf("error setting up test: %v", err)
	}
	mgm, err := biMgm.NewBankIntegrationManagement(biStorage.GetLoggerDBConnection(), biStorage.GetRedisInstance())
	if err != nil {
		t.Fatalf("error initializing management service: %v", err)
	}

	if err = mgm.UpdatePartneredBanks(context.Background(), &bank_integration_models.PartneredBank{
		ID:                 1,
		BankName:           "Bank BCA Edit",
		DefaultPicturePath: "/path/to/new/picture",
		PartnershipStatus:  true,
	}); err != nil {
		t.Fatalf("error updating partnered bank: %v", err)
	}
}

func TestDeletePartneredBanks(t *testing.T) {
	if err := setup(); err != nil {
		t.Fatalf("error setting up test: %v", err)
	}
	mgm, err := biMgm.NewBankIntegrationManagement(biStorage.GetLoggerDBConnection(), biStorage.GetRedisInstance())
	if err != nil {
		t.Fatalf("error initializing management service: %v", err)
	}

	if err = mgm.DeletePartneredBank(context.Background(), 1); err != nil {
		t.Fatalf("error deleting partnered bank: %v", err)
	}
}

func TestEditBankIntegratedFeatures(t *testing.T) {
	if err := setup(); err != nil {
		t.Fatalf("error setting up test: %v", err)
	}
	mgm, err := biMgm.NewBankIntegrationManagement(biStorage.GetLoggerDBConnection(), biStorage.GetRedisInstance())
	if err != nil {
		t.Fatalf("error initializing management service: %v", err)
	}

	if err = mgm.EditBankIntegratedFeatures(context.Background(), []*bank_integration_models.IntegratedFeatureAdd{
		{
			IDBank:        1,
			IDFeature:     1,
			IDFeatureType: 1,
			Note:          "Note 1",
			Status:        "true",
		},
		{
			IDBank:        1,
			IDFeature:     2,
			IDFeatureType: 1,
			Note:          "Note 1",
			Status:        "false",
		},
		{
			IDBank:        1,
			IDFeature:     3,
			IDFeatureType: 1,
			Note:          "Note 1",
			Status:        "true",
		},
	}); err != nil {
		t.Fatalf("error updating integrated bank features: %v", err)
	}
}
