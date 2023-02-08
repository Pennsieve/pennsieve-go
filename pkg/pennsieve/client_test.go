package pennsieve

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
	"time"
)

type ClientTestSuite struct {
	suite.Suite
	MockCognitoServer
	MockPennsieveServer
	TestClient *Client
}

func (s *ClientTestSuite) SetupTest() {
	s.MockCognitoServer = NewMockCognitoServerDefault(s.T())
	s.MockPennsieveServer = NewMockPennsieveServerDefault(s.T())
	AWSEndpoints = AWSCognitoEndpoints{IdentityProviderEndpoint: s.IdProviderServer.URL}
	s.TestClient = NewClient(
		APIParams{ApiHost: s.Server.URL})
}

func (s *ClientTestSuite) TearDownTest() {
	s.MockCognitoServer.Close()
	s.MockPennsieveServer.Close()
	AWSEndpoints.Reset()
}

type requestBody struct {
	Name     string `json:"name"`
	ReqValue int    `json:"req_value"`
}

type responseBody struct {
	Name      string `json:"name"`
	RespValue []int  `json:"resp_value"`
}

func (s *ClientTestSuite) TestSendUnauthenticatedRequest() {
	expectedQueryKey := "query_param_key"
	expectedQueryValue := "query_param_value"
	expectedURL := "/resource-name"

	reqBody := requestBody{
		Name:     "request name",
		ReqValue: 38,
	}

	expectedRespBody := responseBody{
		Name:      "response body",
		RespValue: []int{1, 2, 3},
	}

	s.Mux.HandleFunc(expectedURL, func(writer http.ResponseWriter, request *http.Request) {
		s.Equal("POST", request.Method, "unexpected http method for SendUnauthenticated")
		// Check query params
		s.NoError(request.ParseForm())
		s.Equal(expectedQueryValue, request.Form.Get(expectedQueryKey))
		// Check request body
		actualReqBody := requestBody{}
		s.NoError(json.NewDecoder(request.Body).Decode(&actualReqBody))
		s.Equal(reqBody, actualReqBody)
		// Send expected response body
		respBodyBytes, err := json.Marshal(expectedRespBody)
		s.NoError(err)
		_, err = writer.Write(respBodyBytes)
		s.NoError(err)
	})

	reqBodyBytes, err := json.Marshal(reqBody)
	s.NoError(err)
	reqBodyReader := bytes.NewBuffer(reqBodyBytes)
	requestURL := fmt.Sprintf("%s%s?%s=%s", s.Server.URL, expectedURL, expectedQueryKey, expectedQueryValue)
	request, err := http.NewRequest("POST", requestURL, reqBodyReader)
	s.NoError(err)

	actualRespBody := responseBody{}

	s.NoError(s.TestClient.sendUnauthenticatedRequest(context.Background(), request, &actualRespBody))

	s.Equal(expectedRespBody, actualRespBody)

}

func (s *ClientTestSuite) TestSendRequest() {
	expectedQueryKey := "query_param_key"
	expectedQueryValue := "query_param_value"
	expectedURL := "/authenticated-resource-name"

	reqBody := requestBody{
		Name:     "request name",
		ReqValue: 38,
	}

	expectedRespBody := responseBody{
		Name:      "response body",
		RespValue: []int{1, 2, 3},
	}

	s.Mux.HandleFunc(expectedURL, func(writer http.ResponseWriter, request *http.Request) {
		s.Equal("PUT", request.Method, "unexpected http method for SendRequest")
		// Check query params
		s.NoError(request.ParseForm())
		s.Equal(expectedQueryValue, request.Form.Get(expectedQueryKey))
		// Check request body
		actualReqBody := requestBody{}
		s.NoError(json.NewDecoder(request.Body).Decode(&actualReqBody))
		s.Equal(reqBody, actualReqBody)
		// Check Authentication headers
		s.Equal(s.TestClient.OrganizationNodeId, request.Header.Get("X-ORGANIZATION-ID"))
		expectedAuthHeader := fmt.Sprintf("Bearer %s", s.TestClient.APISession.Token)
		s.Equal(expectedAuthHeader, request.Header.Get("Authorization"))
		// Send expected response body
		respBodyBytes, err := json.Marshal(expectedRespBody)
		s.NoError(err)
		_, err = writer.Write(respBodyBytes)
		s.NoError(err)
	})

	reqBodyBytes, err := json.Marshal(reqBody)
	s.NoError(err)
	reqBodyReader := bytes.NewBuffer(reqBodyBytes)
	requestURL := fmt.Sprintf("%s%s?%s=%s", s.Server.URL, expectedURL, expectedQueryKey, expectedQueryValue)
	request, err := http.NewRequest("PUT", requestURL, reqBodyReader)
	s.NoError(err)

	actualRespBody := responseBody{}

	s.NoError(s.TestClient.sendRequest(context.Background(), request, &actualRespBody))

	s.Equal(expectedRespBody, actualRespBody)

}

func (s *ClientTestSuite) TestSendRequestExpired() {
	expectedURL := "/expired-resource-name"
	s.Mux.HandleFunc(expectedURL, func(writer http.ResponseWriter, request *http.Request) {
		s.Equal("GET", request.Method, "unexpected http method for SendRequestExpired")
		// Check Authentication headers
		s.Equal(s.TestClient.OrganizationNodeId, request.Header.Get("X-ORGANIZATION-ID"))
		expectedAuthHeader := fmt.Sprintf("Bearer %s", s.TestClient.APISession.Token)
		s.Equal(expectedAuthHeader, request.Header.Get("Authorization"))
		// Send response
		respBodyBytes, err := json.Marshal(responseBody{})
		s.NoError(err)
		_, err = writer.Write(respBodyBytes)
		s.NoError(err)
	})

	requestURL := fmt.Sprintf("%s%s", s.Server.URL, expectedURL)
	request, err := http.NewRequest("GET", requestURL, nil)
	s.NoError(err)

	response := responseBody{}

	// Send initial request to initialize APISession and org id
	s.NoError(s.TestClient.sendRequest(context.Background(), request, &response))

	// Session should still be active
	firstAccessToken := s.TestClient.APISession.Token
	s.NoError(s.TestClient.sendRequest(context.Background(), request, &response))
	s.Equal(firstAccessToken, s.TestClient.APISession.Token)

	// Expire session to make sure we get a new one
	s.TestClient.APISession.Expiration = time.Now().Add(-time.Minute * 30).UTC()
	s.NoError(s.TestClient.sendRequest(context.Background(), request, &response))
	s.NotEqual(firstAccessToken, s.TestClient.APISession.Token)
	s.True(s.TestClient.APISession.Expiration.After(time.Now()))

}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
