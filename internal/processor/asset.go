package processor

import (
	"context"
	"xrf197ilz35aq/internal/model"
)

type AssetProcessor interface {
	CreateAsset(ctx context.Context, userCtx model.UserContext, req model.AssetRequest) (bool, error)
}

type assetProcessor struct{}

func (assetProcessor) CreateAsset(ctx context.Context, userCtx model.UserContext, req model.AssetRequest) (bool, error) {
	return true, nil
}

func NewAssetProcessor() AssetProcessor {
	return &assetProcessor{}
}
