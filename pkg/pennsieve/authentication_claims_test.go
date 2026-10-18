package pennsieve

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The ID token is read without verifying its signature, only to take its
// claims (organization, expiry).
func TestIdTokenClaimsAreReadUnverified(t *testing.T) {
	exp := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	idToken := NewTestJWT(t, "N:organization:abcd", "9999", time.Until(exp))

	token, _, err := jwt.NewParser().ParseUnverified(idToken, jwt.MapClaims{})
	require.NoError(t, err)
	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, "N:organization:abcd", claims[OrgNodeIdClaimKey])
	assert.Equal(t, "9999", claims[OrgIdClaimKey])
	assert.WithinDuration(t, exp, getTokenExpFromClaim(claims), time.Second)
}
