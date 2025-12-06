package model

import (
	"fmt"
)

type AssetRequest struct {
	Name  string `json:"name"`
	OrgId string `json:"orgId"`
}

func (req *AssetRequest) Validate() error {
	if req.OrgId == "" {
		return fmt.Errorf("invalid orgId")
	}
	if req.Name == "" || len(req.Name) < 2 || len(req.Name) > 32 {
		return fmt.Errorf("invalid org name")
	}
	return nil
}
