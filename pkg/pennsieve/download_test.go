package pennsieve

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/download"
	"github.com/stretchr/testify/suite"
)

type DownloadServiceTestSuite struct {
	suite.Suite
	API2Server MockPennsieveServer
	MockCognitoServer
	client *Client
}

func (s *DownloadServiceTestSuite) SetupTest() {
	s.MockCognitoServer = NewMockCognitoServerDefault(s.T())
	s.API2Server = NewMockPennsieveServerDefault(s.T())
	AWSEndpoints = AWSCognitoEndpoints{IdentityProviderEndpoint: s.IdProviderServer.URL}
	s.client = NewClient(APIParams{ApiHost2: s.API2Server.Server.URL, ClientName: "pennsieve-agent/1.4.2"})
}

func (s *DownloadServiceTestSuite) TearDownTest() {
	s.API2Server.Close()
	s.MockCognitoServer.Close()
	AWSEndpoints.Reset()
}

// serveManifest answers three pages of two files each, following cursors.
func (s *DownloadServiceTestSuite) serveManifest(requests *[]download.ManifestRequest) {
	s.API2Server.Mux.HandleFunc("/downloads/manifests", func(w http.ResponseWriter, r *http.Request) {
		s.Equal("POST", r.Method)
		s.Equal("N:dataset:1", r.URL.Query().Get("dataset_id"))
		s.Equal("pennsieve-agent/1.4.2", r.Header.Get("X-Pennsieve-Client"))
		var req download.ManifestRequest
		s.NoError(json.NewDecoder(r.Body).Decode(&req))
		*requests = append(*requests, req)

		pages := map[string]download.ManifestPage{
			"":   {Next: "c1"},
			"c1": {Next: "c2"},
			"c2": {},
		}
		page, ok := pages[req.Cursor]
		s.True(ok, "unexpected cursor %q", req.Cursor)
		page.Header = download.ManifestHeader{Count: 6, Size: 600}
		n := len(*requests)
		for i := 0; i < 2; i++ {
			id := int64((n-1)*2 + i + 1)
			page.Data = append(page.Data, download.ManifestFile{
				NodeId: fmt.Sprintf("N:package:%d", id), FileId: id, FileName: fmt.Sprintf("f%d.csv", id),
				Path: []string{"study"}, URL: fmt.Sprintf("https://s3/f%d?sig", id), Size: 100,
			})
		}
		s.NoError(json.NewEncoder(w).Encode(page))
	})
}

func (s *DownloadServiceTestSuite) TestGetManifestPage() {
	var requests []download.ManifestRequest
	s.serveManifest(&requests)
	page, err := s.client.Download.GetManifestPage(context.Background(), "N:dataset:1",
		download.ManifestRequest{NodeIds: []string{"N:collection:1"}, Limit: 2})
	if s.NoError(err) {
		s.Equal(6, page.Header.Count, "the header counts the whole selection")
		s.Len(page.Data, 2)
		s.Equal([]string{"study"}, page.Data[0].Path)
		s.Equal("c1", page.Next)
	}
	s.Equal([]download.ManifestRequest{{NodeIds: []string{"N:collection:1"}, Limit: 2}}, requests)
}

func (s *DownloadServiceTestSuite) TestWalkManifestFollowsEveryPage() {
	var requests []download.ManifestRequest
	s.serveManifest(&requests)
	var ids []int64
	err := s.client.Download.WalkManifest(context.Background(), "N:dataset:1", download.ManifestRequest{},
		func(page *download.ManifestPage) error {
			for _, f := range page.Data {
				ids = append(ids, f.FileId)
			}
			return nil
		})
	s.NoError(err)
	s.Equal([]int64{1, 2, 3, 4, 5, 6}, ids)
	s.Equal([]string{"", "c1", "c2"}, []string{requests[0].Cursor, requests[1].Cursor, requests[2].Cursor})
	s.Empty(requests[0].NodeIds, "no nodeIds: the whole dataset")
}

func (s *DownloadServiceTestSuite) TestWalkManifestStopsOnACallbackError() {
	var requests []download.ManifestRequest
	s.serveManifest(&requests)
	stop := errors.New("disk full")
	err := s.client.Download.WalkManifest(context.Background(), "N:dataset:1", download.ManifestRequest{},
		func(*download.ManifestPage) error { return stop })
	s.ErrorIs(err, stop)
	s.Len(requests, 1, "no further pages requested")
}

func (s *DownloadServiceTestSuite) TestClientHeaderOnlyWhenNamed() {
	s.API2Server.Mux.HandleFunc("/downloads/manifests", func(w http.ResponseWriter, r *http.Request) {
		s.Empty(r.Header.Values("X-Pennsieve-Client"))
		s.NoError(json.NewEncoder(w).Encode(download.ManifestPage{}))
	})
	client := NewClient(APIParams{ApiHost2: s.API2Server.Server.URL})
	_, err := client.Download.GetManifestPage(context.Background(), "N:dataset:1", download.ManifestRequest{})
	s.NoError(err)
}

func (s *DownloadServiceTestSuite) TestUpdateparamsMovesToTheNewHost() {
	var requests []download.ManifestRequest
	s.serveManifest(&requests)
	client := NewClient(APIParams{ApiHost2: "http://127.0.0.1:1", ClientName: "pennsieve-agent/1.4.2"})
	client.Updateparams(APIParams{ApiHost2: s.API2Server.Server.URL, ClientName: "pennsieve-agent/1.4.2"})
	_, err := client.Download.GetManifestPage(context.Background(), "N:dataset:1", download.ManifestRequest{})
	s.NoError(err)
	s.Len(requests, 1)
}

func (s *DownloadServiceTestSuite) TestManifestOfASavedSelection() {
	var requests []download.ManifestRequest
	s.serveManifest(&requests)
	_, err := s.client.Download.GetManifestPage(context.Background(), "N:dataset:1",
		download.ManifestRequest{SelectionId: "sel_aaaaaaaaaaaaaaaaaaaaaaaaaa"})
	s.NoError(err)
	s.Equal([]download.ManifestRequest{{SelectionId: "sel_aaaaaaaaaaaaaaaaaaaaaaaaaa"}}, requests)
}

// servePublicManifest answers two pages of a published selection, recording
// each request's raw JSON.
func (s *DownloadServiceTestSuite) servePublicManifest(bodies *[]map[string]any) {
	s.API2Server.Mux.HandleFunc("/downloads/public/manifests", func(w http.ResponseWriter, r *http.Request) {
		s.Equal("POST", r.Method)
		s.Empty(r.URL.Query(), "published data is not scoped to a workspace or dataset")
		s.Equal("pennsieve-agent/1.4.2", r.Header.Get("X-Pennsieve-Client"))
		var body map[string]any
		s.NoError(json.NewDecoder(r.Body).Decode(&body))
		*bodies = append(*bodies, body)

		page := download.PublicManifestPage{
			Header: download.PublicManifestHeader{DatasetId: 5347, Version: 2, Count: 3, Size: 300},
		}
		switch body["cursor"] {
		case nil:
			page.Data = []download.PublicManifestFile{
				{FileName: "a.csv", Path: []string{"study"}, URL: "https://s3/a?sig", Size: 100, SHA256: "aa"},
				{FileName: "b.csv", Path: []string{"study"}, URL: "https://s3/b?sig", Size: 100},
			}
			page.Next = "p1"
		case "p1":
			page.Data = []download.PublicManifestFile{{FileName: "c.csv", URL: "https://s3/c?sig", Size: 100}}
			page.Skipped = []download.SkippedFile{{FileName: "old.bin", Reason: "no_object_version"}}
		default:
			s.Failf("unexpected cursor", "%v", body["cursor"])
		}
		s.NoError(json.NewEncoder(w).Encode(page))
	})
}

func (s *DownloadServiceTestSuite) TestWalkPublicManifestFollowsEveryPage() {
	var bodies []map[string]any
	s.servePublicManifest(&bodies)
	var names []string
	var skipped []string
	err := s.client.Download.WalkPublicManifest(context.Background(),
		download.PublicManifestRequest{DatasetId: 5347, Paths: []string{"study"}},
		func(page *download.PublicManifestPage) error {
			s.Equal(2, page.Header.Version, "the version resolved for latest")
			for _, f := range page.Data {
				names = append(names, f.FileName)
			}
			for _, f := range page.Skipped {
				skipped = append(skipped, f.FileName+":"+f.Reason)
			}
			return nil
		})
	s.NoError(err)
	s.Equal([]string{"a.csv", "b.csv", "c.csv"}, names)
	s.Equal([]string{"old.bin:no_object_version"}, skipped)
	s.Equal([]map[string]any{
		{"datasetId": float64(5347), "paths": []any{"study"}},
		{"datasetId": float64(5347), "paths": []any{"study"}, "cursor": "p1"},
	}, bodies, "no version is the latest")
}

func (s *DownloadServiceTestSuite) TestPublicManifestOfASavedSelection() {
	var bodies []map[string]any
	s.servePublicManifest(&bodies)
	page, err := s.client.Download.GetPublicManifestPage(context.Background(),
		download.PublicManifestRequest{SelectionId: "sel_bbbbbbbbbbbbbbbbbbbbbbbbbb"})
	if s.NoError(err) {
		s.Equal("aa", page.Data[0].SHA256)
	}
	s.Equal([]map[string]any{{"selectionId": "sel_bbbbbbbbbbbbbbbbbbbbbbbbbb"}}, bodies,
		"only the selection: the service rejects it with a dataset or paths")
}

func (s *DownloadServiceTestSuite) TestGetSelection() {
	s.API2Server.Mux.HandleFunc("/downloads/selections/", func(w http.ResponseWriter, r *http.Request) {
		s.Equal("GET", r.Method)
		switch r.URL.Path {
		case "/downloads/selections/sel_aaaaaaaaaaaaaaaaaaaaaaaaaa":
			_, _ = w.Write([]byte(`{"id":"sel_aaaaaaaaaaaaaaaaaaaaaaaaaa","kind":"workspace","datasetNodeId":"N:dataset:1","expiresAt":"2026-10-06T12:00:00.000Z"}`))
		case "/downloads/selections/sel_bbbbbbbbbbbbbbbbbbbbbbbbbb":
			_, _ = w.Write([]byte(`{"id":"sel_bbbbbbbbbbbbbbbbbbbbbbbbbb","kind":"public","datasetId":5347,"version":2,"expiresAt":"2026-10-06T12:00:00.000Z"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"selection not found, or expired; select the files again"}`))
		}
	})

	sel, err := s.client.Download.GetSelection(context.Background(), "sel_aaaaaaaaaaaaaaaaaaaaaaaaaa")
	if s.NoError(err) {
		s.Equal(download.Selection{
			Id: "sel_aaaaaaaaaaaaaaaaaaaaaaaaaa", Kind: download.SelectionWorkspace,
			DatasetNodeId: "N:dataset:1", ExpiresAt: "2026-10-06T12:00:00.000Z",
		}, *sel)
	}

	sel, err = s.client.Download.GetSelection(context.Background(), "sel_bbbbbbbbbbbbbbbbbbbbbbbbbb")
	if s.NoError(err) {
		s.Equal(download.SelectionPublic, sel.Kind)
		s.Equal(int64(5347), sel.DatasetId)
		s.Equal(2, sel.Version)
	}

	_, err = s.client.Download.GetSelection(context.Background(), "sel_cccccccccccccccccccccccccc")
	var httpErr *HTTPError
	if s.ErrorAs(err, &httpErr) {
		s.Equal(http.StatusNotFound, httpErr.StatusCode)
		s.Contains(httpErr.Message, "select the files again")
	}

	_, err = s.client.Download.GetSelection(context.Background(), "../manifests")
	s.ErrorAs(err, &httpErr, "the id stays one path segment")
}

func TestDownloadService(t *testing.T) {
	suite.Run(t, new(DownloadServiceTestSuite))
}

// Every service follows a profile switch to the new hosts.
func TestUpdateparamsMovesEveryService(t *testing.T) {
	c := NewClient(APIParams{ApiHost: "https://api.old", ApiHost2: "https://api2.old"})
	c.Updateparams(APIParams{ApiHost: "https://api.new", ApiHost2: "https://api2.new"})

	d := c.Dataset.(*datasetService)
	p := c.Package.(*packageService)
	hosts := map[string]string{
		"Dataset":      d.BaseUrl,
		"Dataset2":     d.BaseUrl2,
		"Discover":     c.Discover.(*discoverService).BaseUrl,
		"Timeseries":   c.Timeseries.(*timeseriesService).BaseUrl,
		"Manifest":     c.Manifest.(*manifestService).baseUrl,
		"Account":      c.Account.(*accountService).BaseUrl,
		"Package":      p.baseUrl,
		"Package2":     p.baseUrl2,
		"Download":     c.Download.(*downloadService).BaseUrl,
		"User":         c.User.(*userService).BaseUrl,
		"Organization": c.Organization.(*organizationService).baseUrl,
	}
	for name, host := range hosts {
		if strings.Contains(host, ".old") {
			t.Errorf("%s still calls %s after Updateparams", name, host)
		}
	}
}
