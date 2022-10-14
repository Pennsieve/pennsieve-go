package pennsieve

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/user"
	"net/http"
)

type mockHTTPService struct {
	APISession         APISession
	APICredentials     APICredentials
	OrganizationNodeId string
	OrganizationId     int
}

func (c *mockHTTPService) sendUnauthenticatedRequest(ctx context.Context, req *http.Request, v interface{}) error {
	return c.sendRequest(ctx, req, v)
}

func (c *mockHTTPService) sendRequest(ctx context.Context, req *http.Request, v interface{}) error {

	// Return an object of the requested type.
	switch t := v.(type) {
	case *user.User:
		v.(*user.User).ID = "1234"
	case float64:
		fmt.Println("float64:", v)
	default:
		fmt.Println(t)
	}

	return nil
}

func (c *mockHTTPService) GetCredentials() APICredentials {
	return c.APICredentials
}

func (c *mockHTTPService) SetSession(s APISession) {
	c.APISession = s
}

func (c *mockHTTPService) GetAPICredentials() APICredentials {
	return c.APICredentials
}

func (c *mockHTTPService) SetOrganization(orgId int, orgNodeId string) {
	c.OrganizationId = orgId
	c.OrganizationNodeId = orgNodeId
}
