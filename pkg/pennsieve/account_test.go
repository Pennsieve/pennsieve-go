package pennsieve

import (
	"context"
	"encoding/json"
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

func TestAccountServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AccountServiceTestSuite))
}
