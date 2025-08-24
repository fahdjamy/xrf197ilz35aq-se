package processor

import (
	"context"
	"fmt"
	"net/http"
	v1 "xrf197ilz35aq/gen/account/v1"
	"xrf197ilz35aq/internal"
	"xrf197ilz35aq/internal/model"
)

type AccountProcessor interface {
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
		return model.AccountResponse{}, fmt.Errorf("failed to create account: %w", err)
	}

	return ap.convertAcctResponse(resp.Account, req.Timezone)
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
		return nil, fmt.Errorf("failed to find accounts: %w", err)
	}

	if resp.Accounts == nil || len(resp.Accounts) == 0 {
		return []model.AccountResponse{}, nil
	}

	var accounts []model.AccountResponse
	for _, account := range resp.Accounts {
		response, err := ap.convertAcctResponse(account, "UTC")
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

func (ap *accountProcessor) convertAcctResponse(response *v1.AccountResponse, tx string) (model.AccountResponse, error) {
	createdAt, err := convertTimestamp(response.CreationTime, tx)
	if err := checkConvertedGrpcTimeErr(err); err != nil {
		return model.AccountResponse{}, err
	}
	modifiedAt, err := convertTimestamp(response.ModificationTime, tx)
	if err := checkConvertedGrpcTimeErr(err); err != nil {
		return model.AccountResponse{}, err
	}

	walletHoldingResp := response.WalletHolding
	walletModificationTime, err := convertTimestamp(walletHoldingResp.ModificationTime, tx)
	if err := checkConvertedGrpcTimeErr(err); err != nil {
		return model.AccountResponse{}, err
	}

	return model.AccountResponse{
		Timezone:         tx,
		CreatedAt:        createdAt,
		ModificationTime: modifiedAt,
		Status:           response.Status,
		Locked:           response.Locked,
		AccountId:        response.AccountId,
		Wallet: model.WalletHolding{
			ModificationTime: walletModificationTime,
			Balance:          walletHoldingResp.Balance,
			Currency:         walletHoldingResp.Currency,
		},
	}, nil
}

func NewAccountProcessor(grpcAcctService v1.AccountServiceClient) AccountProcessor {
	return &accountProcessor{grpcAcctClient: grpcAcctService}
}
