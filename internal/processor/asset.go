package processor

import (
	"context"
	"net/http"
	v1 "xrf197ilz35aq/gen/xrfq1/asset/v1"
	"xrf197ilz35aq/internal"
	"xrf197ilz35aq/internal/model"
	"xrf197ilz35aq/internal/service"
)

type AssetProcessor interface {
	CreateAsset(ctx context.Context, userCtx model.UserContext, req model.AssetRequest) (bool, error)
}

type assetProcessor struct {
	grpcAcctClient v1.AssetServiceClient
	orgService     service.OrgService
}

func (ap *assetProcessor) CreateAsset(ctx context.Context, userCtx model.UserContext, req model.AssetRequest) (bool, error) {
	if err := req.Validate(); err != nil {
		return false, &internal.ExternalError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
	}

	gRPCCtxWithHeaders := createGrpcContextWithHeaders(ctx, userCtx)
	_, err := ap.grpcAcctClient.Create(gRPCCtxWithHeaders, &v1.CreateRequest{
		Name:         "",
		Symbol:       "",
		Description:  "",
		Organization: "",
	})

	if err != nil {
		return false, handleGrpcError(err)
	}

	return true, nil
}

func NewAssetProcessor() AssetProcessor {
	return &assetProcessor{}
}
