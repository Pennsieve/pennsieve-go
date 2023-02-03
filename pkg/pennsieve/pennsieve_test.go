package pennsieve

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/authentication"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// This file contains httptest.Servers wrapped for convenience to be used as mocks for
// both AWS' Cognito Identity Provider and Pennsieve API servers.
// It's probably not idiomatic to have these here

// A httptest.Server that mocks Amazon CognitoIdentityProvider
// Create with NewMockCognitoServer*() to get one with the correct handler already in place.
type MockCognitoServer struct {
	IdProviderServer *httptest.Server
}

const (
	OrgNodeIdClaimKey = "custom:organization_node_id"
	OrgIdClaimKey     = "custom:organization_id"
)

// Returns a MockCognitoServer configured to always return an unexpired IDToken JWT that includes the claims made in expectedClaims.
// If expectedClaims is nil, the returned claims will be as in NewMockCognitoServerDefault()
func NewMockCognitoServer(t *testing.T, expectedClaims map[string]any) MockCognitoServer {
	if expectedClaims == nil {
		return NewMockCognitoServerDefault(t)
	}
	counter := 0
	cognitoServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.String() != "/" {
			t.Errorf("unexpected cognito identity provider call: expected: %q, got: %q", "/", request.URL)
		}

		idTokenString := NewTestJWTWithClaims(t, expectedClaims, time.Hour)
		// hack to probably get a unique access token
		counter++
		accessToken := fmt.Sprintf("access-token-%d", counter)
		_, err := fmt.Fprintf(writer, `{"AuthenticationResult": {"AccessToken": %q, "ExpiresIn": 3600, "IdToken": %q, "RefreshToken": %q, "TokenType": "Bearer"}, "ChallengeParameters": {}}`,
			accessToken,
			idTokenString,
			"mock-refresh-token")
		if err != nil {
			t.Error("error writing AuthenticationResult")
		}
	}))

	return MockCognitoServer{IdProviderServer: cognitoServer}
}

// Returns a MockCognitoServer configured to always return a JWT IdToken that includes org node id and org id claims.
// Also, an exp claim with a time an hour in the future.
func NewMockCognitoServerDefault(t *testing.T) MockCognitoServer {
	return NewMockCognitoServer(t, map[string]any{
		OrgNodeIdClaimKey: "N:Organization:abcd",
		OrgIdClaimKey:     "9999",
	})
}

func (m *MockCognitoServer) Close() {
	m.IdProviderServer.Close()
}

func NewTestJWTWithClaims(t *testing.T, customClaims map[string]any, sessionTTL time.Duration) string {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(sessionTTL).UTC().Unix(),
	}
	for k, v := range customClaims {
		claims[k] = v
	}
	idToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	idTokenString, err := idToken.SignedString([]byte("test-signing-key"))
	if err != nil {
		t.Errorf("error getting signed string from JWT token: %s", err)
	}
	return idTokenString
}

func NewTestJWT(t *testing.T, orgNodeId, orgId string, sessionTTL time.Duration) string {
	return NewTestJWTWithClaims(t, map[string]any{
		OrgNodeIdClaimKey: orgNodeId,
		OrgIdClaimKey:     orgId,
	}, sessionTTL)
}

// A httptest.Server intended to mock an external server.
// The intended use is that Mux should be Server's Handler.
// Mux is separated out this way so that tests can add handlers to
// an already started Server by adding to Mux instead of Server.
type MockServer struct {
	Server *httptest.Server
	Mux    *http.ServeMux
}

func (m *MockServer) Close() {
	m.Server.Close()
}

// A httptest.Server intended to mock a Pennsieve API server.
// In tests, Handlers should be added to Mux.
type MockPennsieveServer struct {
	MockServer
}

// Returns a MockPennsieveServer with the same configuration as
// NewMockPennsieveServer() with a default authentication.CognitoConfig
func NewMockPennsieveServerDefault(t *testing.T) MockPennsieveServer {
	return NewMockPennsieveServer(t, authentication.CognitoConfig{
		Region: "us-east-1",
		UserPool: authentication.UserPool{
			Region:      "us-east-1",
			ID:          "mock-user-pool-id",
			AppClientID: "mock-user-pool-app-client-id",
		},
		TokenPool: authentication.TokenPool{
			Region:      "us-east-1",
			AppClientID: "mockTokenPoolAppClientId",
		},
		IdentityPool: authentication.IdentityPool{
			Region: "us-east-1",
			ID:     "mock-identity-pool-id",
		}})
}

// Returns a MockPennsieveServer with a GET "/authentication/cognito-config" handler already configured. It will return the given
// expectedCognitoConfig. If you don't care about the content of the CognitoConfig you can use NewMockPennsieveServerDefault() instead.
// There is also a "/" handler configured that always fails the test to catch any unexpected requests.
func NewMockPennsieveServer(t *testing.T, expectedCognitoConfig authentication.CognitoConfig) MockPennsieveServer {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	mux.HandleFunc("/authentication/cognito-config", func(writer http.ResponseWriter, request *http.Request) {
		body, err := json.Marshal(expectedCognitoConfig)
		if err != nil {
			t.Error("could not marshal mock CognitoConfig")
		}
		_, err = writer.Write(body)
		if err != nil {
			t.Error("error writing CognitoConfig response")
		}
	})
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		t.Errorf("Unhandled request: method: %q, path: %q. If this call is expected add a HandleFunc to MockPennsieveServer.Mux", request.Method, request.URL)
	})
	return MockPennsieveServer{
		MockServer{
			Server: server,
			Mux:    mux,
		},
	}
}
