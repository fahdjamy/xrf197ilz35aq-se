package model

import "errors"

type AccountRequest struct {
	AccountType string `json:"accountType" validate:"required"`
	Currency    string `json:"currency" validate:"required"`
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
	return nil
}

type AccountResponse struct {
	Status    string        `json:"status"`
	Locked    bool          `json:"locked"`
	AccountId string        `json:"accountId"`
	CreatedAt int64         `json:"createdAt"`
	Timezone  string        `json:"timezone"`
	Wallet    WalletHolding `json:"walletHolding"`
}

type WalletHolding struct {
	Balance          float64 `json:"balance"`
	AccountId        string  `json:"accountId"`
	Currency         string  `json:"currency"`
	ModificationTime int64   `json:"modificationTime"`
}

type AccountsRequest struct {
	Accounts []AccountRequest `json:"accounts"`
}
