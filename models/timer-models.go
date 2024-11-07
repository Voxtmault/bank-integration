package bank_integration_models

import "time"

type TimerPayment struct {
	Id        int
	NumVA     string
	IdBank    int
	ExpiredAt time.Time
}
