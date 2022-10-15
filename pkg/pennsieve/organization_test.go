package pennsieve

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Testing very basic component of UserService to get active user.
// We mock the HTTP-client to return a user.
func TestGetOrg(t *testing.T) {

	ht := mockHTTPClient{
		APISession:         APISession{},
		APICredentials:     APICredentials{},
		OrganizationNodeId: "",
		OrganizationId:     0,
	}

	testOrganizationService := NewOrganizationService(&ht, "https://test.com")
	org, _ := testOrganizationService.Get(context.Background(), "N:org:1234")
	assert.Equal(t, "Valid Organization", org.Organization.Name, "Org name must match.")
}

func TestList(t *testing.T) {

	mockClient := mockHTTPClient{
		APISession:         APISession{},
		APICredentials:     APICredentials{},
		OrganizationNodeId: "",
		OrganizationId:     0,
	}

	testOrganizationService := NewOrganizationService(&mockClient, "https://test.com")
	org, _ := testOrganizationService.List(context.Background())

	assert.Equal(t, "First Org", org.Organizations[0].Organization.Name, "Org name must match.")
	assert.Equal(t, "Second Org", org.Organizations[1].Organization.Name, "Org name must match.")

}
