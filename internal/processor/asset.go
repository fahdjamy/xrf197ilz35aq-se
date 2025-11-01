package processor

import (
	"context"
	v1 "xrf197ilz35aq/gen/xrfq1/asset/v1"
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

func (assetProcessor) CreateAsset(ctx context.Context, userCtx model.UserContext, req model.AssetRequest) (bool, error) {
	return true, nil
}

func NewAssetProcessor() AssetProcessor {
	return &assetProcessor{}
}
