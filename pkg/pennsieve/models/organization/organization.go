package organization

import (
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/user"
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
	Orcid                   user.Orcid              `json:"orcid"`
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
	Orcid                   user.Orcid              `json:"orcid"`
}
type Organizations struct {
	Organization   Organization     `json:"organization"`
	IsAdmin        bool             `json:"isAdmin"`
	Owners         []Owners         `json:"owners"`
	Administrators []Administrators `json:"administrators"`
	IsOwner        bool             `json:"isOwner"`
}
