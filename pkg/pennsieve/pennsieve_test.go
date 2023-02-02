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

type MockCognitoServer struct {
	IdProviderServer *httptest.Server
}

func NewMockCognitoServer(t *testing.T, expectedClaims jwt.MapClaims) MockCognitoServer {
	if expectedClaims == nil {
		return NewMockCognitoServerDefault(t)
	}
	cognitoServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.String() != "/" {
			t.Errorf("unexpected cognito identity provider call: expected: %q, got: %q", "/", request.URL)
		}
		idToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expectedClaims)

		idTokenString, err := idToken.SignedString([]byte("test-signing-key"))
		if err != nil {
			t.Errorf("error getting signed string from JWT token: %s", err)
		}
		fmt.Fprintf(writer, `{"AuthenticationResult": {"AccessToken": %q, "ExpiresIn": 3600, "IdToken": %q, "RefreshToken": %q, "TokenType": "Bearer"}, "ChallengeParameters": {}}`,
			"mock-access-token",
			idTokenString,
			"mock-refresh-token")
	}))

	return MockCognitoServer{IdProviderServer: cognitoServer}
}

func NewMockCognitoServerDefault(t *testing.T) MockCognitoServer {
	return NewMockCognitoServer(t, jwt.MapClaims{
		"custom:organization_node_id": "N:Organization:abcd",
		"custom:organization_id":      "9999",
		"exp":                         time.Now().Add(time.Hour).Unix(),
	})
}

func (m *MockCognitoServer) Close() {
	m.IdProviderServer.Close()
}

type MockPennsieveServer struct {
	Server *httptest.Server
	Mux    *http.ServeMux
}

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

func NewMockPennsieveServer(t *testing.T, expectedCognitoConfig authentication.CognitoConfig) MockPennsieveServer {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	mux.HandleFunc("/authentication/cognito-config", func(writer http.ResponseWriter, request *http.Request) {
		body, err := json.Marshal(expectedCognitoConfig)
		if err != nil {
			t.Error("could not marshal mock CognitoConfig")
		}
		writer.Write(body)
	})
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		t.Errorf("Unhandled request: method: %q, path: %q. If this call is expected add a HandleFunc to MockPennsieveServer.Mux", request.Method, request.URL)
	})

	return MockPennsieveServer{Server: server, Mux: mux}
}

func (m *MockPennsieveServer) Close() {
	m.Server.Close()
}
