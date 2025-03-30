package pennsieve

import (
	"context"
	"encoding/json"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest/manifestFile"
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
	AWSEndpoints = AWSCognitoEndpoints{IdentityProviderEndpoint: s.IdProviderServer.URL}
	client := NewClient(APIParams{
		ApiHost:  s.APIServer.Server.URL,
		ApiHost2: s.API2Server.Server.URL,
	})
	s.TestService = client.Manifest
}

func (s *ManifestServiceTestSuite) TearDownTest() {
	s.MockCognitoServer.Close()
	s.APIServer.Close()
	s.API2Server.Close()
	AWSEndpoints.Reset()
}

func (s *ManifestServiceTestSuite) TestCreateManifest() {
	expectedManifestId := "1234"
	expectedManifestNodeId := "N:manifest:1234"
	expectedDatasetId := "N:Dataset:1a2b"
	expectedReqDto := manifest.DTO{
		ID:        expectedManifestId,
		DatasetId: expectedDatasetId,
		Files: []manifestFile.FileDTO{
			{S3Key: "/path/to/one"},
			{S3Key: "/path/to/two"},
		},
		Status: manifest.Initiated,
	}

	s.API2Server.Mux.HandleFunc("/manifest", func(writer http.ResponseWriter, request *http.Request) {
		s.Equal("POST", request.Method, "unexpected http method for Create Manifest")
		err := request.ParseForm()
		if s.NoError(err) {
			s.Equal(expectedDatasetId, request.Form.Get("dataset_id"), "unexpected dataset_id in query param")
		}
		reqDto := manifest.DTO{}
		err = json.NewDecoder(request.Body).Decode(&reqDto)
		if s.NoError(err) {
			s.Equal(expectedReqDto, reqDto, "unexpected Create reqeust body")
		}
		respObj := manifest.PostResponse{
			ManifestNodeId: expectedManifestNodeId,
		}
		respBody, err := json.Marshal(respObj)
		if s.NoError(err) {
			_, err := writer.Write(respBody)
			s.NoError(err)
		}
	})
	postResponse, err := s.TestService.Create(context.Background(), expectedReqDto)
	if s.NoError(err) {
		s.Equal(expectedManifestNodeId, postResponse.ManifestNodeId, "unexpected manifest node id")
	}
}

func TestManifestService(t *testing.T) {
	suite.Run(t, new(ManifestServiceTestSuite))
}
