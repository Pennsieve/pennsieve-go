package pennsieve

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/manifest"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/dataset"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/organization"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/user"
	"net/http"
)

type mockHTTPClient struct {
	APISession         APISession
	APICredentials     APICredentials
	OrganizationNodeId string
	OrganizationId     int
}

func (c *mockHTTPClient) sendUnauthenticatedRequest(ctx context.Context, req *http.Request, v interface{}) error {
	return c.sendRequest(ctx, req, v)
}

func (c *mockHTTPClient) sendRequest(ctx context.Context, req *http.Request, v interface{}) error {

	// Return an object of the requested type.
	switch t := v.(type) {
	case *user.User:
		v.(*user.User).ID = "1234"
	case *manifest.PostResponse:
		v.(*manifest.PostResponse).ManifestNodeId = "1234"
	case *dataset.GetDatasetResponse:
		v.(*dataset.GetDatasetResponse).Content.Name = "Test Dataset"
	case *organization.GetOrganizationResponse:
		v.(*organization.GetOrganizationResponse).Organization.Name = "Valid Organization"
	case *organization.GetOrganizationsResponse:
		orgs := []organization.Organizations{
			{Organization: organization.Organization{Name: "First Org"}},
			{Organization: organization.Organization{Name: "Second Org"}},
		}
		v.(*organization.GetOrganizationsResponse).Organizations = orgs
	default:
		fmt.Println(t)
	}

	return nil
}

func (c *mockHTTPClient) GetCredentials() APICredentials {
	return c.APICredentials
}

func (c *mockHTTPClient) SetSession(s APISession) {
	c.APISession = s
}

func (c *mockHTTPClient) GetAPICredentials() APICredentials {
	return c.APICredentials
}

func (c *mockHTTPClient) SetOrganization(orgId int, orgNodeId string) {
	c.OrganizationId = orgId
	c.OrganizationNodeId = orgNodeId
}

func (c *mockHTTPClient) SetBasePathForServices(baseUrlV1 string, baseUrlV2 string) {

}
