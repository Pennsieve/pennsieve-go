package pennsieve

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/download"
)

// DownloadService calls download-service (api2 /downloads), which signs
// download links for dataset files and records each download.
type DownloadService interface {
	// GetManifestPage returns one page of signed URLs for a selection.
	GetManifestPage(ctx context.Context, datasetId string, req download.ManifestRequest) (*download.ManifestPage, error)
	// WalkManifest requests every page of a selection in turn, calling fn
	// with each as it arrives: the links are signed just before they're
	// used, and the whole selection is never held in memory. An error from
	// fn stops the walk and is returned.
	WalkManifest(ctx context.Context, datasetId string, req download.ManifestRequest, fn func(*download.ManifestPage) error) error
	SetBaseUrl(url string)
}

type downloadService struct {
	client  PennsieveHTTPClient
	BaseUrl string
}

func NewDownloadService(client PennsieveHTTPClient, baseUrl string) *downloadService {
	return &downloadService{client: client, BaseUrl: baseUrl}
}

func (s *downloadService) SetBaseUrl(url string) {
	s.BaseUrl = url
}

func (s *downloadService) GetManifestPage(ctx context.Context, datasetId string, manifestReq download.ManifestRequest) (*download.ManifestPage, error) {
	body, err := json.Marshal(manifestReq)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("dataset_id", datasetId)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/downloads/manifests?%s", s.BaseUrl, params.Encode()), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = req.Context()
	}
	page := download.ManifestPage{}
	if err := s.client.sendRequest(ctx, req, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

func (s *downloadService) WalkManifest(ctx context.Context, datasetId string, manifestReq download.ManifestRequest, fn func(*download.ManifestPage) error) error {
	for {
		page, err := s.GetManifestPage(ctx, datasetId, manifestReq)
		if err != nil {
			return err
		}
		if err := fn(page); err != nil {
			return err
		}
		if page.Next == "" {
			return nil
		}
		manifestReq.Cursor = page.Next
	}
}
