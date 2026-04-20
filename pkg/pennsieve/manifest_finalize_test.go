package pennsieve

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FinalizeManifestFilesTestSuite struct {
	suite.Suite
	MockCognitoServer
	APIServer   MockPennsieveServer
	API2Server  MockPennsieveServer
	TestService ManifestService
}

func (s *FinalizeManifestFilesTestSuite) SetupTest() {
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

func (s *FinalizeManifestFilesTestSuite) TearDownTest() {
	s.MockCognitoServer.Close()
	s.APIServer.Close()
	s.API2Server.Close()
	AWSEndpoints.Reset()
}

// TestFinalizeDefaultOmitsOnConflict asserts that the default call (no
// FinalizeOption) omits the onConflict field from the request body —
// legacy clients must hit the server with exactly the same bytes they did
// before this SDK version.
func (s *FinalizeManifestFilesTestSuite) TestFinalizeDefaultOmitsOnConflict() {
	s.API2Server.Mux.HandleFunc("/upload/manifest/files/finalize", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		s.NoError(json.Unmarshal(body, &got))
		_, present := got["onConflict"]
		s.False(present, "onConflict must be omitted from the default-call body (omitempty)")
		w.Write([]byte(`{"results":[]}`))
	})

	_, err := s.TestService.FinalizeManifestFiles(
		context.Background(),
		"N:Dataset:abc",
		"N:Manifest:m1",
		[]FinalizeFile{{UploadID: "u1", Size: 10, SHA256: "sha"}},
	)
	s.NoError(err)
}

// TestFinalizeWithOnConflict asserts that WithOnConflict propagates into
// the JSON body so the upload-service-v2 finalize handler can consume it.
func (s *FinalizeManifestFilesTestSuite) TestFinalizeWithOnConflict() {
	s.API2Server.Mux.HandleFunc("/upload/manifest/files/finalize", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got struct {
			ManifestNodeID string `json:"manifestNodeId"`
			OnConflict     string `json:"onConflict"`
		}
		s.NoError(json.Unmarshal(body, &got))
		s.Equal("N:Manifest:m1", got.ManifestNodeID)
		s.Equal(FinalizeOnConflictReplace, got.OnConflict, "WithOnConflict must populate the body field")
		w.Write([]byte(`{"results":[]}`))
	})

	_, err := s.TestService.FinalizeManifestFiles(
		context.Background(),
		"N:Dataset:abc",
		"N:Manifest:m1",
		[]FinalizeFile{{UploadID: "u1", Size: 10, SHA256: "sha"}},
		WithOnConflict(FinalizeOnConflictReplace),
	)
	s.NoError(err)
}

// TestFinalizeOptionConstants assert the constants match the server-side
// accepted values. If upload-service-v2's validOnConflict switch changes,
// this should be updated in lockstep.
func TestFinalizeOptionConstants(t *testing.T) {
	assert.Equal(t, "keepBoth", FinalizeOnConflictKeepBoth)
	assert.Equal(t, "replace", FinalizeOnConflictReplace)
}

func TestFinalizeManifestFilesService(t *testing.T) {
	suite.Run(t, new(FinalizeManifestFilesTestSuite))
}