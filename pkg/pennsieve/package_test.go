package pennsieve

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/ps_package"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PackageServiceTestSuite struct {
	suite.Suite
	MockPennsieveServer
	MockCognitoServer
	TestService PackageService
}

func (s *PackageServiceTestSuite) SetupTest() {
	s.MockCognitoServer = NewMockCognitoServerDefault(s.T())
	s.MockPennsieveServer = NewMockPennsieveServerDefault(s.T())
	AWSEndpoints = AWSCognitoEndpoints{IdentityProviderEndpoint: s.IdProviderServer.URL}
	client := NewClient(APIParams{
		ApiHost: s.Server.URL,
	})
	s.TestService = client.Package
}

func (s *PackageServiceTestSuite) TearDownTest() {
	s.MockPennsieveServer.Close()
	s.MockCognitoServer.Close()
	AWSEndpoints.Reset()
}

//TestGetPresignedUrl tests GET presigned url. This also invokes the GET Package Sources endpoint.
func (s *PackageServiceTestSuite) TestGetPresignedUrl() {
	expectedPackageId := "N:Package:1234"
	expectedName := "Test Package"
	expectedUrl := "https://1234"
	expectedFileId := 123
	expectedFileName := "File1.txt"

	expectedPath := fmt.Sprintf("/packages/%s/sources-paged", expectedPackageId)
	s.Mux.HandleFunc(expectedPath, func(writer http.ResponseWriter, request *http.Request) {
		s.Equalf("GET", request.Method, "expected method GET, got: %q", request.Method)
		packageResponse := ps_package.GetPackageSourcesResponse{
			Limit:  0,
			Offset: 0,
			Results: []ps_package.Result{
				{
					Content: ps_package.Content{
						PackageID: "1234",
						Name:      expectedName,
						Filename:  expectedFileName,
						S3Bucket:  "s3://testBucket",
						S3Key:     "s3://testBucket/testFile",
						ID:        int64(expectedFileId),
					},
				},
			},
			TotalCount: 0,
		}
		body, err := json.Marshal(packageResponse)
		if s.NoError(err) {
			_, err := writer.Write(body)
			s.NoError(err)
		}
	})

	expectedPath2 := fmt.Sprintf("/packages/%s/files/%d", expectedPackageId, expectedFileId)
	s.Mux.HandleFunc(expectedPath2, func(writer http.ResponseWriter, request *http.Request) {
		s.Equalf("GET", request.Method, "expected method GET, got: %q", request.Method)
		packageResponse := ps_package.GetPresignedUrlDTO{URL: expectedUrl}
		body, err := json.Marshal(packageResponse)
		if s.NoError(err) {
			_, err := writer.Write(body)
			s.NoError(err)
		}
	})

	getPackageResp, err := s.TestService.GetPresignedUrl(context.Background(), expectedPackageId, false)
	if s.NoError(err) {
		s.Equal(expectedFileName, getPackageResp.Files[0].Name, "unexpected name")
		s.Equal(expectedUrl, getPackageResp.Files[0].URL, "unexpected url")
	}
}

func TestPackageServiceSuite(t *testing.T) {
	suite.Run(t, new(PackageServiceTestSuite))
}
