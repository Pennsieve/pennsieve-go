package pennsieve

import (
	"context"
	"encoding/json"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/manifest"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type ManifestServiceTestSuite struct {
	suite.Suite
	MockCognitoServer
	APIServer   MockPennsieveServer
	API2Server  MockPennsieveServer
	TestService ManifestService
}

func (s *ManifestServiceTestSuite) SetupTest() {
	s.MockCognitoServer = NewMockCognitoServerDefault(s.T())
	s.APIServer = NewMockPennsieveServerDefault(s.T())
	s.API2Server = NewMockPennsieveServerDefault(s.T())
	client := NewClient(APIParams{
		ApiHost:  s.APIServer.Server.URL,
		ApiHost2: s.API2Server.Server.URL,
	}, s.IdProviderServer.URL)
	s.TestService = client.Manifest
}

func (s *ManifestServiceTestSuite) TearDownTest() {
	s.MockCognitoServer.Close()
	s.APIServer.Close()
	s.API2Server.Close()
}

func (s *ManifestServiceTestSuite) TestCreateManifest() {
	expectedNodeId := "1234"
	expectedDatasetId := "N:Dataset:1a2b"
	s.API2Server.Mux.HandleFunc("/manifest/", func(writer http.ResponseWriter, request *http.Request) {
		s.Equal("POST", request.Method, "unexpected http method for Create Manifest")
		respObj := manifest.PostResponse{
			ManifestNodeId: expectedNodeId,
		}
		respBody, err := json.Marshal(respObj)
		if s.NoError(err) {
			writer.Write(respBody)
		}
	})
	manifest, _ := s.TestService.Create(context.Background(), manifest.DTO{
		ID:        expectedNodeId,
		DatasetId: expectedDatasetId,
		Files:     nil,
		Status:    0,
	})
	s.Equal(expectedNodeId, manifest.ManifestNodeId, "unexpected manifest node id")
}

func TestManifestService(t *testing.T) {
	suite.Run(t, new(ManifestServiceTestSuite))
}
