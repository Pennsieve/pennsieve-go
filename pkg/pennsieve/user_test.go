package pennsieve

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Testing very basic component of UserService to get active user.
// We mock the HTTP-client to return a user.
func TestGetUser(t *testing.T) {

	client := NewClient(APIParams{}, "")

	testUserService := client.User

	user, _ := testUserService.GetUser(context.Background())

	assert.Equal(t, "1234", user.ID, "UserID must match.")

}
