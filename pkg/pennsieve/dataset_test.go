package pennsieve

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/dataset"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type DatasetServiceTestSuite struct {
	suite.Suite
	MockPennsieveServer
	MockCognitoServer
	TestService DatasetService
}

func (s *DatasetServiceTestSuite) SetupTest() {
	s.MockCognitoServer = NewMockCognitoServerDefault(s.T())
	s.MockPennsieveServer = NewMockPennsieveServerDefault(s.T())
	client := NewClient(APIParams{
		ApiHost: s.Server.URL,
	}, &AWSCognitoEndpoints{IdentityProviderEndpoint: s.IdProviderServer.URL})
	s.TestService = client.Dataset
}

func (s *DatasetServiceTestSuite) TearDownTest() {
	s.MockPennsieveServer.Close()
	s.MockCognitoServer.Close()
}

func (s *DatasetServiceTestSuite) TestGetDataset() {
	expectedDatasetId := "N:Dataset:1234"
	expectedDatasetName := "Test Dataset"
	expectedPath := "/datasets/" + expectedDatasetId
	s.Mux.HandleFunc(expectedPath, func(writer http.ResponseWriter, request *http.Request) {
		s.Equalf("GET", request.Method, "expected method GET, got: %q", request.Method)
		datasetResponse := dataset.GetDatasetResponse{Content: dataset.Content{Name: expectedDatasetName}}
		body, err := json.Marshal(datasetResponse)
		if s.NoError(err) {
			_, err := writer.Write(body)
			s.NoError(err)
		}
	})

	getDatasetResp, err := s.TestService.Get(context.Background(), expectedDatasetId)
	if s.NoError(err) {
		s.Equal(expectedDatasetName, getDatasetResp.Content.Name, "unexpected name")
	}
}

func (s *DatasetServiceTestSuite) TestCreateDataset() {
	expectedName := "created dataset"
	expectedDescription := "testing create dataset"
	expectedTags := []string{"tag one", "tag-2"}
	expectedTagString := fmt.Sprintf("[%q, %q]", expectedTags[0], expectedTags[1])
	s.Mux.HandleFunc("/datasets/", func(writer http.ResponseWriter, request *http.Request) {
		s.Equalf("POST", request.Method, "expected method POST, got: %q", request.Method)
		requestBody := struct {
			Name        string
			Description string
			Tags        []string
		}{}
		err := json.NewDecoder(request.Body).Decode(&requestBody)
		if err != nil {
			s.Fail("error decoding create body request", err)
		}
		s.Equal(expectedName, requestBody.Name, "unexpected name in request body")
		s.Equal(expectedDescription, requestBody.Description, "unexpected description in request body")
		s.Equal(expectedTags, requestBody.Tags, "unexpected tags in request body")
		resp := dataset.CreateDatasetResponse{
			Content: dataset.Content{
				Name:        expectedName,
				Description: expectedDescription,
				Tags:        expectedTags,
			},
		}
		respBody, err := json.Marshal(resp)
		if s.NoError(err) {
			_, err := writer.Write(respBody)
			s.NoError(err)

		}
	})
	resp, err := s.TestService.Create(context.Background(), expectedName, expectedDescription, expectedTagString)
	if s.NoError(err) {
		s.Equal(expectedName, resp.Content.Name, "unexpected name")
		s.Equal(expectedDescription, resp.Content.Description, "unexpected description")
		s.Equal(expectedTags, resp.Content.Tags, "unexpected tags")
	}
}

func TestDatasetServiceSuite(t *testing.T) {
	suite.Run(t, new(DatasetServiceTestSuite))
}
