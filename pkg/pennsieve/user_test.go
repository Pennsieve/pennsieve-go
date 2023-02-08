package pennsieve

import (
	"context"
	"encoding/json"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/user"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type UserServiceTestSuite struct {
	suite.Suite
	MockCognitoServer
	MockPennsieveServer
	TestService UserService
}

func (s *UserServiceTestSuite) SetupTest() {
	s.MockCognitoServer = NewMockCognitoServerDefault(s.T())
	s.MockPennsieveServer = NewMockPennsieveServerDefault(s.T())
	AWSEndpoints = AWSCognitoEndpoints{IdentityProviderEndpoint: s.IdProviderServer.URL}
	client := NewClient(APIParams{
		ApiHost: s.Server.URL,
	})
	s.TestService = client.User
}

func (s *UserServiceTestSuite) TearDownTest() {
	s.MockCognitoServer.Close()
	s.MockPennsieveServer.Close()
	AWSEndpoints.Reset()
}

// Testing very basic component of UserService to get active user.
func (s *UserServiceTestSuite) TestGetUser() {
	expectedUserId := "1234"
	s.Mux.HandleFunc("/user/", func(writer http.ResponseWriter, request *http.Request) {
		s.Equal("GET", request.Method, "unexpected http method called for GetUser")
		respUser := user.User{
			ID: expectedUserId,
		}
		respBody, err := json.Marshal(respUser)
		if s.NoError(err) {
			_, err = writer.Write(respBody)
			s.NoError(err)
		}
	})

	userResp, err := s.TestService.GetUser(context.Background())
	if s.NoError(err) {
		s.Equal(expectedUserId, userResp.ID, "UserID must match.")
	}

}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
