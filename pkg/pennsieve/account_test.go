package pennsieve

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/account"
	"github.com/stretchr/testify/suite"
)

type AccountServiceTestSuite struct {
	suite.Suite
	MockPennsieveServer
	MockCognitoServer
	TestService AccountService
}

func (s *AccountServiceTestSuite) SetupTest() {
	s.MockCognitoServer = NewMockCognitoServerDefault(s.T())
	s.MockPennsieveServer = NewMockPennsieveServerDefault(s.T())
	AWSEndpoints = AWSCognitoEndpoints{IdentityProviderEndpoint: s.IdProviderServer.URL}
	client := NewClient(APIParams{
		ApiHost: s.Server.URL,
	})
	s.TestService = client.Account
}

func (s *AccountServiceTestSuite) TearDownTest() {
	s.MockPennsieveServer.Close()
	s.MockCognitoServer.Close()
	AWSEndpoints.Reset()
}

func (s *AccountServiceTestSuite) TestGetPennsieveAccounts() {
	expectedAccountId := "12345"
	expectedAccountType := "aws"
	expectedPath := "/pennsieve-accounts/" + expectedAccountType
	s.Mux.HandleFunc(expectedPath, func(writer http.ResponseWriter, request *http.Request) {
		s.Equalf("GET", request.Method, "expected method GET, got: %q", request.Method)
		pennsieveAccountsResponse := account.GetPennsieveAccountsResponse{AccountId: expectedAccountId, AccountType: expectedAccountType}
		body, err := json.Marshal(pennsieveAccountsResponse)
		if s.NoError(err) {
			_, err := writer.Write(body)
			s.NoError(err)
		}
	})

	getPennsieveAccountsResp, err := s.TestService.GetPennsieveAccounts(context.Background(), expectedAccountType)
	if s.NoError(err) {
		s.Equal(expectedAccountId, getPennsieveAccountsResp.AccountId, "unexpected accountId")
	}
}

func (s *AccountServiceTestSuite) TestCreateAccount() {
	expectedAccountId := "xyz"
	expectedAccountType := "aws"
	expectedRoleName := "SomeRole"
	expectedExternalId := "SomeExternalId"
	expectedUuid := "SomeRandomID"

	s.Mux.HandleFunc("/accounts", func(writer http.ResponseWriter, request *http.Request) {
		s.Equalf("POST", request.Method, "expected method POST, got: %q", request.Method)
		requestBody := struct {
			AccountId   string `json:"accountId"`
			AccountType string `json:"accountType"`
			RoleName    string `json:"roleName"`
			ExternalId  string `json:"externalId"`
		}{}
		data, _ := io.ReadAll(request.Body)
		err := json.Unmarshal(data, &requestBody)
		if err != nil {
			s.Fail("error decoding create body request", err)
		}
		s.Equal(expectedAccountId, requestBody.AccountId, "unexpected accountId in request body")
		s.Equal(expectedAccountType, requestBody.AccountType, "unexpected account type in request body")
		s.Equal(expectedRoleName, requestBody.RoleName, "unexpected role name in request body")
		s.Equal(expectedExternalId, requestBody.ExternalId, "unexpected external ID in request body")
		resp := account.CreateAccountResponse{
			Uuid: expectedUuid,
		}
		respBody, err := json.Marshal(resp)
		if s.NoError(err) {
			_, err := writer.Write(respBody)
			s.NoError(err)

		}
	})
	resp, err := s.TestService.CreateAccount(context.Background(), expectedAccountId, expectedAccountType, expectedRoleName, expectedExternalId)
	if s.NoError(err) {
		s.Equal(expectedUuid, resp.Uuid, "unexpected uuid in request body")
	}
}

func TestAccountServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AccountServiceTestSuite))
}
