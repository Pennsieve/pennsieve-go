package pennsieve

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/discover"
	"github.com/stretchr/testify/suite"
)

type DiscoverServiceTestSuite struct {
	suite.Suite
	MockPennsieveServer
	MockCognitoServer
	TestService DiscoverService
}

func (s *DiscoverServiceTestSuite) SetupTest() {
	s.MockCognitoServer = NewMockCognitoServerDefault(s.T())
	s.MockPennsieveServer = NewMockPennsieveServerDefault(s.T())
	AWSEndpoints = AWSCognitoEndpoints{IdentityProviderEndpoint: s.IdProviderServer.URL}
	client := NewClient(APIParams{
		ApiHost: s.Server.URL,
	})
	s.TestService = client.Discover
}

func (s *DiscoverServiceTestSuite) TearDownTest() {
	s.MockPennsieveServer.Close()
	s.MockCognitoServer.Close()
	AWSEndpoints.Reset()
}

func (s *DiscoverServiceTestSuite) TestGetDatasetByVersion() {
	expectedDatasetId := int32(5069)
	expectedVersionId := int32(2)
	expectedPath := fmt.Sprintf("/discover/datasets/%v/versions/%v", expectedDatasetId, expectedVersionId)
	s.Mux.HandleFunc(expectedPath, func(writer http.ResponseWriter, request *http.Request) {
		s.Equalf("GET", request.Method, "expected method GET, got: %q", request.Method)
		datasetByVersionResponse := discover.GetDatasetByVersionResponse{
			ID: expectedDatasetId, Version: expectedVersionId, Uri: "s3://pennsieve-dev-discover-publish50-use1/5069/"}
		body, err := json.Marshal(datasetByVersionResponse)
		if s.NoError(err) {
			_, err := writer.Write(body)
			s.NoError(err)
		}
	})

	getDatasetResp, err := s.TestService.GetDatasetByVersion(
		context.Background(), expectedDatasetId, expectedVersionId)
	if s.NoError(err) {
		s.Equal(expectedDatasetId, getDatasetResp.ID, "unexpected ID")
	}
}

func (s *DiscoverServiceTestSuite) TestGetDatasetMetadataByVersion() {
	expectedDatasetId := int32(5069)
	expectedVersionId := int32(2)
	expectedPath := fmt.Sprintf("/discover/datasets/%v/versions/%v/metadata", expectedDatasetId, expectedVersionId)
	s.Mux.HandleFunc(expectedPath, func(writer http.ResponseWriter, request *http.Request) {
		s.Equalf("GET", request.Method, "expected method GET, got: %q", request.Method)
		datasetByVersionResponse := discover.GetDatasetMetadataByVersionResponse{
			ID: expectedDatasetId, Version: expectedVersionId, Files: []discover.DatasetFile{{Name: "test.txt"}}}
		body, err := json.Marshal(datasetByVersionResponse)
		if s.NoError(err) {
			_, err := writer.Write(body)
			s.NoError(err)
		}
	})

	getDatasetResp, err := s.TestService.GetDatasetMetadataByVersion(
		context.Background(), expectedDatasetId, expectedVersionId)
	if s.NoError(err) {
		s.Equal(expectedDatasetId, getDatasetResp.ID, "unexpected ID")
	}
}

func (s *DiscoverServiceTestSuite) TestGetDatasetFileByVersion() {
	expectedDatasetId := int32(5069)
	expectedVersionId := int32(2)
	expectedFilename := "manifest.json"
	expectedPath := fmt.Sprintf("/discover/datasets/%v/versions/%v/files",
		expectedDatasetId, expectedVersionId)
	s.Mux.HandleFunc(expectedPath, func(writer http.ResponseWriter, request *http.Request) {
		s.Equalf("GET", request.Method, "expected method GET, got: %q", request.Method)
		datasetFileByVersionResponse := discover.GetDatasetFileByVersionResponse{
			Name: expectedFilename, Uri: "s3://pennsieve-dev-discover-publish50-use1/5069/manifest.json"}
		body, err := json.Marshal(datasetFileByVersionResponse)
		if s.NoError(err) {
			_, err := writer.Write(body)
			s.NoError(err)
		}
	})

	getDatasetResp, err := s.TestService.GetDatasetFileByVersion(
		context.Background(), expectedDatasetId, expectedVersionId, expectedFilename)
	if s.NoError(err) {
		s.Equal(expectedFilename, getDatasetResp.Name, "unexpected Name")
	}
}

func TestDiscoverServiceSuite(t *testing.T) {
	suite.Run(t, new(DiscoverServiceTestSuite))
}
