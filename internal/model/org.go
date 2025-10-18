package model

import "time"

type OrgDetails struct {
	OrgId        string    `json:"orgId"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created"`
	Category     string    `json:"category"`
	UpdatedAt    time.Time `json:"lastUpdated"`
	Description  string    `json:"description"`
	MembersCount int       `json:"membersCount"`
	IsAnonymous  bool      `json:"isAnonymous"`
}
