package pennsieve

import (
	"context"
	"encoding/json"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/organization"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type OrganizationServiceTestSuite struct {
	suite.Suite
	MockCognitoServer
	MockPennsieveServer
	TestService OrganizationService
}

func (s *OrganizationServiceTestSuite) SetupTest() {
	s.MockCognitoServer = NewMockCognitoServerDefault(s.T())
	s.MockPennsieveServer = NewMockPennsieveServerDefault(s.T())
	AWSEndpoints = AWSCognitoEndpoints{IdentityProviderEndpoint: s.IdProviderServer.URL}
	client := NewClient(
		APIParams{ApiHost: s.Server.URL})
	s.TestService = client.Organization
}

func (s *OrganizationServiceTestSuite) TearDownTest() {
	s.MockCognitoServer.Close()
	s.MockPennsieveServer.Close()
	AWSEndpoints.Reset()
}

// Testing very basic component of UserService to get active user.
func (s *OrganizationServiceTestSuite) TestGetOrg() {
	expectedOrgId := "N:org:1234"
	expectedOrgName := "Valid Organization"
	s.Mux.HandleFunc("/organizations/"+expectedOrgId, func(writer http.ResponseWriter, request *http.Request) {
		s.Equal("GET", request.Method, "unexpected http method for GetOrg")
		respOrg := organization.GetOrganizationResponse{
			Organization: organization.Organization{
				ID:   expectedOrgId,
				Name: expectedOrgName},
		}
		respBody, err := json.Marshal(respOrg)
		if s.NoError(err) {
			_, err := writer.Write(respBody)
			s.NoError(err)
		}
	})
	org, err := s.TestService.Get(context.Background(), expectedOrgId)
	if s.NoError(err) {
		s.Equal(expectedOrgId, org.Organization.ID, "unexpected org id")
		s.Equal(expectedOrgName, org.Organization.Name, "unexpected org name")
	}
}

func (s *OrganizationServiceTestSuite) TestList() {
	expectedOrg1 := organization.Organization{
		ID:   "N:org:1234",
		Name: "First Org",
	}
	expectedOrg2 := organization.Organization{
		ID:   "N:org:abcd",
		Name: "Second Org",
	}
	s.Mux.HandleFunc("/organizations/", func(writer http.ResponseWriter, request *http.Request) {
		s.Equal("GET", request.Method, "unexpected http method for List")
		respList := organization.GetOrganizationsResponse{
			Organizations: []organization.Organizations{
				{Organization: expectedOrg1},
				{Organization: expectedOrg2},
			},
		}
		respBody, err := json.Marshal(respList)
		if s.NoError(err) {
			_, err := writer.Write(respBody)
			s.NoError(err)
		}
	})

	org, err := s.TestService.List(context.Background())
	if s.NoError(err) {
		s.Len(org.Organizations, 2)
		s.Equal(expectedOrg1, org.Organizations[0].Organization)
		s.Equal(expectedOrg2, org.Organizations[1].Organization)
	}
}

func TestOrganizationService(t *testing.T) {
	suite.Run(t, new(OrganizationServiceTestSuite))
}
