package model

import (
	"errors"
	"fmt"
	"time"
)

type AccountRequest struct {
	Currency    string `json:"currency" validate:"required"`
	Timezone    string `json:"timezone" validate:"required"`
	AccountType string `json:"accountType" validate:"required"`
}

func (m *AccountRequest) Validate() error {
	acceptedTypes := map[string]bool{
		"SAVING": true,
		"ESCROW": true,
	}
	if !acceptedTypes[m.AccountType] {
		return errors.New("invalid accountType")
	}
	acceptedCurrency := map[string]bool{
		"BTC": true,
		"ETH": true,
		"LTC": true,
		"XRP": true,
	}
	if !acceptedCurrency[m.Currency] {
		return errors.New("invalid currency")
	}
	if m.Timezone == "" {
		return errors.New("timezone is required")
	}
	_, err := time.LoadLocation(m.Timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone: %s", m.Timezone)
	}
	return nil
}

type AccountResponse struct {
	Status           string        `json:"status"`
	Locked           bool          `json:"locked"`
	AccountId        string        `json:"accountId"`
	CreatedAt        time.Time     `json:"createdAt"`
	Timezone         string        `json:"timezone"`
	ModificationTime time.Time     `json:"modificationTime"`
	Wallet           WalletHolding `json:"walletHolding"`
}

type WalletHolding struct {
	Balance          float64   `json:"balance"`
	Currency         string    `json:"currency"`
	ModificationTime time.Time `json:"modificationTime"`
}

type AccountsRequest struct {
	Accounts []AccountRequest `json:"accounts"`
}
