package pennsieve

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/golang-jwt/jwt"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/authentication"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/dataset"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type ServerTestSuite struct {
	suite.Suite
	Server        *httptest.Server
	Mux           *http.ServeMux
	CognitoServer *httptest.Server
}

func (s *ServerTestSuite) SetupTest() {
	s.Mux = http.NewServeMux()
	s.Server = httptest.NewServer(s.Mux)
	s.Mux.HandleFunc("/authentication/cognito-config", func(writer http.ResponseWriter, request *http.Request) {
		cognitoConfig := authentication.CognitoConfig{
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
			},
		}
		body, err := json.Marshal(cognitoConfig)
		if err != nil {
			s.FailNow("could not marshal mock CognitoConfig")
		}
		writer.Write(body)
	})

	s.CognitoServer = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.String() != "/" {
			s.FailNowf("unexpected cognito identity provider call", "expected: %q, got: %q", "/", request.URL)
		}
		body := request.Body
		defer body.Close()
		requestPayload := cognitoidentityprovider.InitiateAuthInput{}
		decodeError := json.NewDecoder(body).Decode(&requestPayload)
		if decodeError != nil {
			s.FailNow("failed to decode cognito input: %s", decodeError)
		}
		fmt.Println(requestPayload)
		idToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"custom:organization_node_id": "N:Organization:abcd",
			"custom:organization_id":      "9999",
			"exp":                         time.Now().Add(time.Hour).Unix(),
		})

		idTokenString, err := idToken.SignedString([]byte("test-signing-key"))
		if err != nil {
			s.FailNowf("error getting signed string from JWT token", "%s", err)
		}
		fmt.Fprintf(writer, `{"AuthenticationResult": {"AccessToken": %q, "ExpiresIn": 3600, "IdToken": %q, "RefreshToken": %q, "TokenType": "Bearer"}, "ChallengeParameters": {}}`,
			"mock-access-token",
			idTokenString,
			"mock-refresh-token")
	}))

}

func (s *ServerTestSuite) TeardownTest() {
	s.Server.Close()
	s.CognitoServer.Close()
}

func (s *ServerTestSuite) TestGetDataset() {
	expectedDatasetId := "N:Dataset:1234"
	expectedDatasetName := "Test Dataset"
	expectedPath := "/datasets/" + expectedDatasetId
	s.Mux.HandleFunc(expectedPath, func(writer http.ResponseWriter, request *http.Request) {
		datasetResponse := dataset.GetDatasetResponse{Content: dataset.Content{Name: expectedDatasetName}}
		body, err := json.Marshal(datasetResponse)
		if err != nil {
			s.FailNow("error marshalling GetDatasetResponse")
		}
		writer.Write(body)

	})
	s.Mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		s.FailNowf("unexpected request", "got %s, expected %s", request.URL, expectedPath)
	})
	client := NewClient(APIParams{
		ApiHost: s.Server.URL,
	}, s.CognitoServer.URL)

	testDatasetService := client.Dataset
	dataset, _ := testDatasetService.Get(context.Background(), expectedDatasetId)
	s.Equal(expectedDatasetName, dataset.Content.Name, "Dataset name must match.")
}

func TestDatasetSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
