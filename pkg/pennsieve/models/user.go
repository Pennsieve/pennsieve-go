package models

import "time"

type User struct {
	ID                      string         `json:"id"`
	Email                   string         `json:"email"`
	FirstName               string         `json:"firstName"`
	MiddleInitial           string         `json:"middleInitial"`
	LastName                string         `json:"lastName"`
	Degree                  string         `json:"degree"`
	Credential              string         `json:"credential"`
	Color                   string         `json:"color"`
	URL                     string         `json:"url"`
	AuthyID                 int            `json:"authyId"`
	IsSuperAdmin            bool           `json:"isSuperAdmin"`
	IsIntegrationUser       bool           `json:"isIntegrationUser"`
	CreatedAt               time.Time      `json:"createdAt"`
	UpdatedAt               time.Time      `json:"updatedAt"`
	PreferredOrganization   string         `json:"preferredOrganization"`
	Orcid                   Orcid          `json:"orcid"`
	PennsieveTermsOfService TermsOfService `json:"pennsieveTermsOfService"`
	CustomTermsOfService    []interface{}  `json:"customTermsOfService"`
	IntID                   int            `json:"intId"`
}

type Orcid struct {
	Name  string `json:"name"`
	Orcid string `json:"orcid"`
}

type TermsOfService struct {
	Version string `json:"version"`
}

type UserOptions struct{}
