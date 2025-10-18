package service

import (
	"context"
	v1 "xrf197ilz35aq/gen/xrfq3/account/v1"
	"xrf197ilz35aq/internal/client"
	"xrf197ilz35aq/internal/model"
)

type OrgService struct {
	apiClient client.ApiClient
}

func (srvc *OrgService) OrgDetails(ctx context.Context, orgId string) (model.OrgDetails, error) {

	return model.OrgDetails{}, nil
}

func NewOrgService(grpcAcctClient v1.AccountServiceClient) OrgService {
	return OrgService{}
}
