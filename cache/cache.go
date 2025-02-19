package bank_integration_cache

import (
	"context"
	"log/slog"

	"github.com/rotisserie/eris"
	biStorage "github.com/voxtmault/bank-integration/storage"
	biUtil "github.com/voxtmault/bank-integration/utils"
)

// Helpers
var (
	PartneredBanksMap   = make(map[string]string)
	BankFeatureTypesMap = make(map[string]string)
	BankFeaturesMap     = make(map[string]string)
	PaymentMethodsMap   = make(map[string]string)
)

func InitCache(ctx context.Context) error {
	rCon := biStorage.GetRedisInstance().RDB

	result, err := rCon.HGetAll(ctx, "partnered-banks").Result()
	if err != nil {
		return eris.Wrap(err, "failed to get partnered_banks from redis")
	}
	for key, value := range result {
		PartneredBanksMap[key] = value
	}

	result, err = rCon.HGetAll(ctx, "bank_feature_types").Result()
	if err != nil {
		return eris.Wrap(err, "failed to get bank_feature_types from redis")
	}
	for key, value := range result {
		BankFeatureTypesMap[key] = value
	}

	result, err = rCon.HGetAll(ctx, "bank_features").Result()
	if err != nil {
		return eris.Wrap(err, "failed to get bank_features from redis")
	}
	for key, value := range result {
		BankFeaturesMap[key] = value
	}

	result, err = rCon.HGetAll(ctx, "payment_methods").Result()
	if err != nil {
		return eris.Wrap(err, "failed to get payment_methods from redis")
	}
	for key, value := range result {
		PaymentMethodsMap[key] = value
	}

	return nil
}

func MessageBrokerSubscribe(ctx context.Context) error {
	rCon := biStorage.GetRedisInstance().RDB

	// Create the subscription variable
	subs := rCon.Subscribe(ctx, biUtil.HelperUpdate)
	defer subs.Close()

	// Validate the subscription
	_, err := subs.Receive(ctx)
	if err != nil {
		return eris.Wrap(err, "failed to subscribe to event")
	}

	// Upon successful subscription, create the listen channel
	ch := subs.Channel()

	// Loop through the channel to listen for messages
	for range ch {
		slog.Debug("received message from message broker, renewing cache")
		// TODO : Diversify the logic, instead of getting the whole cache, only get the updated cache
		InitCache(ctx)
	}

	return nil
}
