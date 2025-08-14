package processor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
	"xrf197ilz35aq/internal"
	"xrf197ilz35aq/internal/model"
	v1 "xrf197ilz35aq/proto/gen/proto/account/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type AccountProcessor interface {
	CreateAccount(ctx context.Context, req model.AccountRequest) (model.AccountResponse, error)
}

type accountProcessor struct {
	grpcAcctClient v1.AccountServiceClient
}

func (ap *accountProcessor) CreateAccount(ctx context.Context, req model.AccountRequest) (model.AccountResponse, error) {
	if err := req.Validate(); err != nil {
		return model.AccountResponse{}, &internal.ExternalError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
	}

	tz := req.Timezone
	if tz == "" {
		tz = "UTC"
	}

	// make a call to the gRPC service to create an account
	resp, err := ap.grpcAcctClient.CreateAccount(ctx, &v1.CreateAccountRequest{
		Currency: req.Currency,
		AcctType: req.AccountType,
		Timezone: "UTC",
	})
	if err != nil {
		return model.AccountResponse{}, fmt.Errorf("failed to create account: %w", err)
	}

	createdAt, err := ConvertTimestamp(resp.Account.CreationTime, tz)
	if err != nil {
		return model.AccountResponse{}, fmt.Errorf("failed to create account: %w", err)
	}
	modifiedAt, err := ConvertTimestamp(resp.Account.ModificationTime, tz)
	if err != nil {
		if errors.Is(err, internal.ErrInvalidGRPCTimeStamp) {
			return model.AccountResponse{}, &internal.ServerError{
				Message: "something went wrong",
				Err:     err,
			}
		}
		return model.AccountResponse{}, &internal.ExternalError{
			Message: err.Error(),
			Code:    400,
		}
	}
	return model.AccountResponse{
		Timezone:         tz,
		CreatedAt:        createdAt,
		ModificationTime: modifiedAt,
		Status:           resp.Account.Status,
		Locked:           resp.Account.Locked,
		AccountId:        resp.Account.AccountId,
		Wallet:           model.WalletHolding{},
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

func NewAccountProcessor(grpcAcctService v1.AccountServiceClient) AccountProcessor {
	return &accountProcessor{grpcAcctClient: grpcAcctService}
}
