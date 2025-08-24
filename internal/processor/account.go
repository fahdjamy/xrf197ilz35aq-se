package processor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
	v1 "xrf197ilz35aq/gen/account/v1"
	"xrf197ilz35aq/internal"
	"xrf197ilz35aq/internal/model"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AccountProcessor interface {
	CreateAccount(ctx context.Context, userCtx model.UserContext, req model.AccountRequest) (model.AccountResponse, error)
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

	// Create gRPC metadata (headers).
	md := metadata.New(map[string]string{
		internal.XrfUserFingerPrint: userCtx.Fingerprint,
	})

	// Create a new context with the metadata attached.
	gRPCCtxWithHeaders := metadata.NewOutgoingContext(ctx, md)

	// make a call to the gRPC service to create an account
	resp, err := ap.grpcAcctClient.CreateAccount(gRPCCtxWithHeaders, &v1.CreateAccountRequest{
		Currency: req.Currency,
		AcctType: req.AccountType,
		Timezone: "UTC",
	})
	if err != nil {
		return model.AccountResponse{}, fmt.Errorf("failed to create account: %w", err)
	}

	createdAt, err := ConvertTimestamp(resp.Account.CreationTime, req.Timezone)
	if err := CheckConvertedGrpcTimeErr(err); err != nil {
		return model.AccountResponse{}, err
	}
	modifiedAt, err := ConvertTimestamp(resp.Account.ModificationTime, req.Timezone)
	if err := CheckConvertedGrpcTimeErr(err); err != nil {
		return model.AccountResponse{}, err
	}

	walletHoldingResp := resp.Account.WalletHolding
	walletModificationTime, err := ConvertTimestamp(walletHoldingResp.ModificationTime, req.Timezone)
	if err := CheckConvertedGrpcTimeErr(err); err != nil {
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

// ConvertTimestamp dynamically converts a protobuf timestamp to a Go time.Time
// in the specified timezone.
func ConvertTimestamp(ts *timestamppb.Timestamp, tz string) (time.Time, error) {
	// check if the timestamp is valid.
	if err := ts.CheckValid(); err != nil {
		return time.Time{}, internal.ErrInvalidGRPCTimeStamp
	}

	// Load the specified location/timezone.
	location, err := time.LoadLocation(tz)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not load location '%s': %w", tz, err)
	}

	// Convert the protobuf timestamp to a Go time.Time object.
	// The AsTime() method returns the time in UTC.
	goTime := ts.AsTime()

	// Return the time in the desired location.
	return goTime.In(location), nil
}

func CheckConvertedGrpcTimeErr(err error) error {
	if err != nil {
		if errors.Is(err, internal.ErrInvalidGRPCTimeStamp) {
			return &internal.ServerError{
				Message: "something went wrong",
				Err:     err,
			}
		}
		return &internal.ExternalError{
			Message: err.Error(),
			Code:    400,
		}
	}
	return nil
}

func NewAccountProcessor(grpcAcctService v1.AccountServiceClient) AccountProcessor {
	return &accountProcessor{grpcAcctClient: grpcAcctService}
}
