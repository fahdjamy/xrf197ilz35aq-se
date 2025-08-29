package processor

import (
	"context"
	"net/http"
	v1 "xrf197ilz35aq/gen/account/v1"
	"xrf197ilz35aq/internal"
	"xrf197ilz35aq/internal/model"
)

type AccountProcessor interface {
	FindAccountByID(ctx context.Context, userCtx model.UserContext, acctId string) (model.AccountResponse, error)
	CreateAccount(ctx context.Context, userCtx model.UserContext, req model.AccountRequest) (model.AccountResponse, error)
	FindAccounts(ctx context.Context, userCtx model.UserContext, req model.FindAccountRequest) ([]model.AccountResponse, error)
}

type accountProcessor struct {
	grpcAcctClient v1.AccountServiceClient
}

func (ap *accountProcessor) CreateAccount(
	ctx context.Context, userCtx model.UserContext,
	req model.AccountRequest) (model.AccountResponse, error) {

	if err := req.Validate(); err != nil {
		return model.AccountResponse{}, &internal.ExternalError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
	}

	gRPCCtxWithHeaders := createGrpcContextWithHeaders(ctx, userCtx)

	// make a call to the gRPC service to create an account
	resp, err := ap.grpcAcctClient.CreateAccount(gRPCCtxWithHeaders, &v1.CreateAccountRequest{
		Currency: req.Currency,
		AcctType: req.AccountType,
		Timezone: "UTC",
	})
	if err != nil {
		return model.AccountResponse{}, handleGrpcError(err)
	}

	return convertAcctResponse(resp.Account, req.Timezone)
}

func (ap *accountProcessor) FindAccounts(ctx context.Context, userCtx model.UserContext,
	req model.FindAccountRequest) ([]model.AccountResponse, error) {
	if err := req.Validate(); err != nil {
		return []model.AccountResponse{}, &internal.ExternalError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
	}

	gRPCCtxWithHeaders := createGrpcContextWithHeaders(ctx, userCtx)
	resp, err := ap.grpcAcctClient.FindAccountsByCurrencyOrType(gRPCCtxWithHeaders, &v1.FindAccountsByCurrencyOrTypeRequest{
		Currencies: &v1.CurrencyList{Currencies: req.Currencies},
		AcctTypes:  &v1.AccountTypesList{Types: req.AccountTypes},
	})

	if err != nil {
		return nil, handleGrpcError(err)
	}

	if resp.Accounts == nil || len(resp.Accounts) == 0 {
		return []model.AccountResponse{}, nil
	}

	var accounts []model.AccountResponse
	for _, account := range resp.Accounts {
		response, err := convertAcctResponse(account, "UTC")
		if err != nil {
			return []model.AccountResponse{}, &internal.ServerError{
				Message: err.Error(),
				Err:     err,
			}
		}
		accounts = append(accounts, response)
	}

	return accounts, nil
}

func (ap *accountProcessor) FindAccountByID(ctx context.Context, userCtx model.UserContext, acctId string) (model.AccountResponse, error) {

	gRPCCtxWithHeaders := createGrpcContextWithHeaders(ctx, userCtx)

	resp, err := ap.grpcAcctClient.FindAccountById(gRPCCtxWithHeaders, &v1.FindAccountByIdRequest{
		AccountId:      acctId,
		IncludeWallets: false,
	})
	if err != nil {
		return model.AccountResponse{}, handleGrpcError(err)
	}

	if resp.Account == nil {
		return model.AccountResponse{}, &internal.ExternalError{
			Message: "Account not found",
			Code:    http.StatusNotFound,
		}
	}

	return convertAcctResponse(resp.Account, "UTC")
}

func convertAcctResponse(response *v1.AccountResponse, timezone string) (model.AccountResponse, error) {
	createdAt, err := convertTimestamp(response.CreationTime, timezone)
	if err := checkConvertedGrpcTimeErr(err); err != nil {
		return model.AccountResponse{}, err
	}
	modifiedAt, err := convertTimestamp(response.ModificationTime, timezone)
	if err := checkConvertedGrpcTimeErr(err); err != nil {
		return model.AccountResponse{}, err
	}

	walletHoldingResp := response.Wallets
	if err := checkConvertedGrpcTimeErr(err); err != nil {
		return model.AccountResponse{}, err
	}

	convertedWallets := make([]model.WalletHolding, len(walletHoldingResp))

	for _, wallet := range walletHoldingResp {
		convertedWallet, err := convertWalletResponse(wallet, timezone)
		if err != nil {
			return model.AccountResponse{}, err
		}
		convertedWallets = append(convertedWallets, convertedWallet)
	}

	return model.AccountResponse{
		Timezone:         timezone,
		CreatedAt:        createdAt,
		ModificationTime: modifiedAt,
		Status:           response.Status,
		Locked:           response.Locked,
		Wallets:          convertedWallets,
		AccountId:        response.AccountId,
	}, nil
}

func convertWalletResponse(wallet *v1.WalletResponse, timezone string) (model.WalletHolding, error) {
	walletModificationTime, err := convertTimestamp(wallet.ModificationTime, timezone)
	if err := checkConvertedGrpcTimeErr(err); err != nil {
		return model.WalletHolding{}, err
	}

	return model.WalletHolding{
		Balance:          wallet.Balance,
		Currency:         wallet.Currency,
		ModificationTime: walletModificationTime,
	}, nil
}

func NewAccountProcessor(grpcAcctService v1.AccountServiceClient) AccountProcessor {
	return &accountProcessor{grpcAcctClient: grpcAcctService}
}
