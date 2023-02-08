package pennsieve

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/authentication"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type AuthenticationServiceTestSuite struct {
	suite.Suite
	MockPennsieveServer
	// Not using MockCognitoServer since we want to add assertions
	// in this suite
	MockCognito MockServer
	TestClient  *Client
	TestService AuthenticationService
}

var expectedCognitoConfig = authentication.CognitoConfig{
	Region: "us-east-1",
	UserPool: authentication.UserPool{
		Region:      "us-east-1",
		ID:          "auth-test-user-pool-id",
		AppClientID: "auth-test-user-pool-app-client-id",
	},
	TokenPool: authentication.TokenPool{
		Region:      "us-east-1",
		AppClientID: "authTestTokenPoolAppClientId",
	},
	IdentityPool: authentication.IdentityPool{
		Region: "us-east-1",
		ID:     "auth-test-identity-pool-id",
	}}

func (s *AuthenticationServiceTestSuite) SetupTest() {
	cognitoMux := http.NewServeMux()
	cognitoServer := httptest.NewServer(cognitoMux)
	s.MockCognito = MockServer{
		Server: cognitoServer,
		Mux:    cognitoMux,
	}
	s.MockPennsieveServer = NewMockPennsieveServer(s.T(), expectedCognitoConfig)
	AWSEndpoints = AWSCognitoEndpoints{IdentityProviderEndpoint: s.MockCognito.Server.URL}
	s.TestClient = NewClient(
		APIParams{ApiHost: s.Server.URL})
	s.TestService = s.TestClient.Authentication
}

func (s *AuthenticationServiceTestSuite) TearDownTest() {
	s.MockCognito.Close()
	s.MockPennsieveServer.Close()
	AWSEndpoints.Reset()
}

func (s *AuthenticationServiceTestSuite) TestGetCognitoConfig() {
	actualConfig, err := s.TestService.getCognitoConfig()
	if s.NoError(err) {
		s.Equal(expectedCognitoConfig, *actualConfig)
	}
}

func (s *AuthenticationServiceTestSuite) TestAuthenticate() {
	expectedApiKey := "api-key"
	expectedApiSecret := "api-secret"

	expectedAccessToken := "auth-test-access-token"
	expectedRefreshToken := "auth-test-refresh-token"
	expectedOrgNodeId := "N:organization:a9b8c7"
	expectedOrgId := "456"
	expectedIdToken := NewTestJWT(s.T(), expectedOrgNodeId, expectedOrgId, time.Hour)

	s.MockCognito.Mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		s.Equal("POST", request.Method, "unexpected http method for cognito authenticate")
		reqMap := map[string]any{}
		if s.NoError(json.NewDecoder(request.Body).Decode(&reqMap)) {
			s.Equal(types.AuthFlowTypeUserPasswordAuth, types.AuthFlowType(reqMap["AuthFlow"].(string)))
			s.Equal(expectedCognitoConfig.TokenPool.AppClientID, reqMap["ClientId"])
			authParams := reqMap["AuthParameters"].(map[string]any)
			s.Equal(expectedApiKey, authParams["USERNAME"])
			s.Equal(expectedApiSecret, authParams["PASSWORD"])
		}
		authenticationResult := types.AuthenticationResultType{
			AccessToken:  &expectedAccessToken,
			IdToken:      &expectedIdToken,
			RefreshToken: &expectedRefreshToken,
		}
		respObj := cognitoidentityprovider.InitiateAuthOutput{
			AuthenticationResult: &authenticationResult,
		}
		respBytes, err := json.Marshal(respObj)
		if s.NoError(err) {
			_, err = writer.Write(respBytes)
			s.NoError(err)
		}
	})

	actualSession, err := s.TestService.Authenticate(expectedApiKey, expectedApiSecret)
	if s.NoError(err) {
		s.Equal(expectedAccessToken, actualSession.Token)
		s.Equal(expectedIdToken, actualSession.IdToken)
		s.Equal(expectedRefreshToken, actualSession.RefreshToken)
		s.True(actualSession.Expiration.After(time.Now()))
	}

	s.Equal(expectedOrgNodeId, s.TestClient.OrganizationNodeId)
	s.Equal(expectedOrgId, fmt.Sprint(s.TestClient.OrganizationId))
	s.Equal(*actualSession, s.TestClient.APISession)

}

func TestAuthenticationServiceSuite(t *testing.T) {
	suite.Run(t, new(AuthenticationServiceTestSuite))
}

// Tests that in default deployment case we are not introducing a custom
// endpoint resolver for AWS
func TestProdEndpointResolver(t *testing.T) {
	client := noOpPennsieveClient{}
	service := NewAuthenticationService(client, "https://example.com")
	//goland:noinspection GoDeprecation
	assert.Nil(t, (*service).awsConfig.EndpointResolver)
	assert.Nil(t, (*service).awsConfig.EndpointResolverWithOptions)
}

type noOpPennsieveClient struct{}

func (noOpPennsieveClient) sendUnauthenticatedRequest(ctx context.Context, req *http.Request, v interface{}) error {
	panic("implement me")
}

func (noOpPennsieveClient) sendRequest(ctx context.Context, req *http.Request, v interface{}) error {
	panic("implement me")
}

func (noOpPennsieveClient) GetAPIParams() *APIParams {
	panic("implement me")
}

func (noOpPennsieveClient) SetSession(s APISession) {
	panic("implement me")
}

func (noOpPennsieveClient) SetOrganization(orgId int, orgNodeId string) {
	panic("implement me")
}

func (noOpPennsieveClient) Updateparams(params APIParams) {
	panic("implement me")
}
