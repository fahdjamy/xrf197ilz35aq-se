package processor

import (
	"context"
	"errors"
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

	createdAt, err := convertTimestamp(resp.Account.CreationTime, req.Timezone)
	if err := checkConvertedGrpcTimeErr(err); err != nil {
		return model.AccountResponse{}, err
	}
	modifiedAt, err := convertTimestamp(resp.Account.ModificationTime, req.Timezone)
	if err := checkConvertedGrpcTimeErr(err); err != nil {
		return model.AccountResponse{}, err
	}

	walletHoldingResp := resp.Account.WalletHolding
	walletModificationTime, err := convertTimestamp(walletHoldingResp.ModificationTime, req.Timezone)
	if err := checkConvertedGrpcTimeErr(err); err != nil {
		return model.AccountResponse{}, err
	}

	return model.AccountResponse{
		CreatedAt:        createdAt,
		ModificationTime: modifiedAt,
		Timezone:         req.Timezone,
		Status:           resp.Account.Status,
		Locked:           resp.Account.Locked,
		AccountId:        resp.Account.AccountId,
		Wallet: model.WalletHolding{
			ModificationTime: walletModificationTime,
			Balance:          walletHoldingResp.Balance,
			Currency:         walletHoldingResp.Currency,
		},
	}, nil
}

func (ap *accountProcessor) FindAccounts(ctx context.Context, userCtx model.UserContext,
	req model.FindAccountRequest) ([]model.AccountResponse, error) {
	if err := req.Validate(); err != nil {
		return []model.AccountResponse{}, &internal.ExternalError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
	}

	return nil, errors.New("not implemented")
}

func NewAccountProcessor(grpcAcctService v1.AccountServiceClient) AccountProcessor {
	return &accountProcessor{grpcAcctClient: grpcAcctService}
}
