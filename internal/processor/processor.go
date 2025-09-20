package processor

import (
	"context"
	"errors"
	"fmt"
	"time"
	"xrf197ilz35aq/internal"
	"xrf197ilz35aq/internal/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Processors struct {
	UserProcessor    UserProcessor
	AuthProcessor    AuthProcessor
	AssetProcessor   AssetProcessor
	AccountProcessor AccountProcessor
}

// convertTimestamp dynamically converts a protobuf timestamp to a Go time.Time
// in the specified timezone.
func convertTimestamp(ts *timestamppb.Timestamp, tz string) (time.Time, error) {
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

func checkConvertedGrpcTimeErr(err error) error {
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

func createGrpcContextWithHeaders(ctx context.Context, userCtx model.UserContext) context.Context {
	// Create gRPC metadata (headers).
	md := metadata.New(map[string]string{
		internal.XrfUserFingerPrint: userCtx.Fingerprint,
	})

	// Create a new context with the metadata attached.
	gRPCCtxWithHeaders := metadata.NewOutgoingContext(ctx, md)
	return gRPCCtxWithHeaders
}

func handleGrpcError(err error) error {
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			// handle gRPC errors
			if st.Code() == codes.NotFound {
				return &internal.ExternalError{Code: 404, Message: st.Message()}
			} else if st.Code() == codes.AlreadyExists {
				return &internal.ExternalError{Code: 409, Message: st.Message()}
			} else if st.Code() == codes.InvalidArgument {
				return &internal.ExternalError{Code: 400, Message: st.Message()}
			} else {
				return &internal.APIClientError{Message: st.Message(), Code: 502}
			}
		}
	}

	return err
}
