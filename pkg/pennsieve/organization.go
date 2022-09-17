package pennsieve

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type GetOrganizationsResponse struct {
	Organizations []Organizations `json:"organizations"`
}
type GetOrganizationResponse struct {
	Organization Organization `json:"organization"`
}

type CustomTermsOfService struct {
	Version        string `json:"version"`
	OrganizationID string `json:"organizationId"`
}
type SubscriptionState struct {
}
type Organization struct {
	Name                 string               `json:"name"`
	EncryptionKeyID      string               `json:"encryptionKeyId"`
	Features             []string             `json:"features"`
	IntID                int                  `json:"intId"`
	Slug                 string               `json:"slug"`
	ID                   string               `json:"id"`
	CustomTermsOfService CustomTermsOfService `json:"customTermsOfService"`
	Storage              int                  `json:"storage"`
	SubscriptionState    SubscriptionState    `json:"subscriptionState"`
	Terms                string               `json:"terms"`
}
type Degree struct {
	EntryName                          string `json:"entryName"`
	EnumeratumEnumEntryStableEntryName string `json:"enumeratum$EnumEntry$$stableEntryName"`
}
type Role struct {
}
type PennsieveTermsOfService struct {
	Version string `json:"version"`
}
type Owners struct {
	UpdatedAt               time.Time               `json:"updatedAt"`
	Email                   string                  `json:"email"`
	Degree                  Degree                  `json:"degree"`
	Role                    Role                    `json:"role"`
	URL                     string                  `json:"url"`
	IsSuperAdmin            bool                    `json:"isSuperAdmin"`
	IsPublisher             bool                    `json:"isPublisher"`
	IntID                   int                     `json:"intId"`
	Color                   string                  `json:"color"`
	PennsieveTermsOfService PennsieveTermsOfService `json:"pennsieveTermsOfService"`
	LastName                string                  `json:"lastName"`
	IsOwner                 bool                    `json:"isOwner"`
	FirstName               string                  `json:"firstName"`
	ID                      string                  `json:"id"`
	AuthyID                 int                     `json:"authyId"`
	CustomTermsOfService    []CustomTermsOfService  `json:"customTermsOfService"`
	MiddleInitial           string                  `json:"middleInitial"`
	PreferredOrganization   string                  `json:"preferredOrganization"`
	CreatedAt               time.Time               `json:"createdAt"`
	Storage                 int                     `json:"storage"`
	Credential              string                  `json:"credential"`
	Orcid                   Orcid                   `json:"orcid"`
}
type Administrators struct {
	UpdatedAt               time.Time               `json:"updatedAt"`
	Email                   string                  `json:"email"`
	Degree                  Degree                  `json:"degree"`
	Role                    Role                    `json:"role"`
	URL                     string                  `json:"url"`
	IsSuperAdmin            bool                    `json:"isSuperAdmin"`
	IsPublisher             bool                    `json:"isPublisher"`
	IntID                   int                     `json:"intId"`
	Color                   string                  `json:"color"`
	PennsieveTermsOfService PennsieveTermsOfService `json:"pennsieveTermsOfService"`
	LastName                string                  `json:"lastName"`
	IsOwner                 bool                    `json:"isOwner"`
	FirstName               string                  `json:"firstName"`
	ID                      string                  `json:"id"`
	AuthyID                 int                     `json:"authyId"`
	CustomTermsOfService    []CustomTermsOfService  `json:"customTermsOfService"`
	MiddleInitial           string                  `json:"middleInitial"`
	PreferredOrganization   string                  `json:"preferredOrganization"`
	CreatedAt               time.Time               `json:"createdAt"`
	Storage                 int                     `json:"storage"`
	Credential              string                  `json:"credential"`
	Orcid                   Orcid                   `json:"orcid"`
}
type Organizations struct {
	Organization   Organization     `json:"organization"`
	IsAdmin        bool             `json:"isAdmin"`
	Owners         []Owners         `json:"owners"`
	Administrators []Administrators `json:"administrators"`
	IsOwner        bool             `json:"isOwner"`
}

type OrganizationService struct {
	client  *Client
	baseUrl string
}

// List lists all the organizations that the user belongs to.
func (o *OrganizationService) List(ctx context.Context) (*GetOrganizationsResponse, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/organizatinons", o.baseUrl), nil)
	if err != nil {
		return nil, err
	}

	res := GetOrganizationsResponse{}
	if err := o.client.SendRequest(ctx, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Get returns a single organization by id.
func (o *OrganizationService) Get(ctx context.Context, id string) (*GetOrganizationResponse, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/organizations/%s", o.baseUrl, id), nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := GetOrganizationResponse{}
	if err := o.client.SendRequest(ctx, req, &res); err != nil {

		fmt.Println("ERROR: ", err)
		return nil, err
	}

	return &res, nil
}
