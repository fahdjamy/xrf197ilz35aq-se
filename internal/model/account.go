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

type UpdateAccountRequest struct {
	Currency    string `json:"currency"`
	Timezone    string `json:"timezone"`
	AccountType string `json:"accountType"`
}

func (m *AccountRequest) Validate() error {
	if m.Timezone == "" {
		// set default timezone to UTC if no timezone is set
		m.Timezone = "UTC"
	}

	acceptedTypes := map[string]bool{
		"Normal": true,
		"Escrow": true,
		"ESCROW": true,
	}
	if !acceptedTypes[m.AccountType] {
		return errors.New("invalid accountType")
	}
	acceptedCurrency := map[string]bool{
		"BTC":  true,
		"ETH":  true,
		"LTC":  true,
		"XRP":  true,
		"USD":  true,
		"XRFQ": true,
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

type FindAccountRequest struct {
	Currencies   []string `json:"currencies"`
	AccountTypes []string `json:"accountTypes"`
}

func (m *FindAccountRequest) Validate() error {
	if m.Currencies == nil {
		m.Currencies = []string{}
	}
	if m.AccountTypes == nil {
		m.AccountTypes = []string{}
	}
	if len(m.Currencies) == 0 && len(m.AccountTypes) == 0 {
		return errors.New("at least one currency or accountType must be provided")
	}
	return nil
}

type AccountResponse struct {
	Status           string          `json:"status"`
	Locked           bool            `json:"locked"`
	AccountId        string          `json:"accountId"`
	CreatedAt        time.Time       `json:"createdAt"`
	Timezone         string          `json:"timezone"`
	Wallets          []WalletHolding `json:"wallets"`
	ModificationTime time.Time       `json:"modificationTime"`
}

type WalletHolding struct {
	Balance          float32   `json:"balance"`
	Currency         string    `json:"currency"`
	ModificationTime time.Time `json:"modificationTime"`
}

type AccountsRequest struct {
	Accounts []AccountRequest `json:"accounts"`
}
