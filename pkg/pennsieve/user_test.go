package pennsieve

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Testing very basic component of UserService to get active user.
// We mock the HTTP-client to return a user.
func TestGetUser(t *testing.T) {

	ht := mockHTTPClient{
		APISession:         APISession{},
		APICredentials:     APICredentials{},
		OrganizationNodeId: "",
		OrganizationId:     0,
	}

	testUserService := NewUserService(&ht, "http://test.com")

	user, _ := testUserService.GetUser(context.Background())

	assert.Equal(t, "1234", user.ID, "UserID must match.")

}
