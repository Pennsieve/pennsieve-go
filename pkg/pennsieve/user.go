package pennsieve

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

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

type UserService struct {
	client  *Client
	baseUrl string
}

func (c *UserService) GetUser(ctx context.Context, options *UserOptions) (*User, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/user", c.baseUrl), nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := User{}
	if err := c.client.sendRequest(ctx, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
