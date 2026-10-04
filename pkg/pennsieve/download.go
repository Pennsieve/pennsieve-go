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
	// GetPublicManifestPage returns one page of signed URLs for a published
	// (Discover) selection. Signed in, embargoed versions the user may
	// preview are included; the user's daily allowance for published data
	// applies.
	GetPublicManifestPage(ctx context.Context, req download.PublicManifestRequest) (*download.PublicManifestPage, error)
	// WalkPublicManifest requests every page of a published selection in
	// turn, as WalkManifest does.
	WalkPublicManifest(ctx context.Context, req download.PublicManifestRequest, fn func(*download.PublicManifestPage) error) error
	// GetSelection says where to download a saved selection from: its kind
	// and dataset. An unknown or expired id is an *HTTPError with status 404.
	GetSelection(ctx context.Context, id string) (*download.Selection, error)
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

func (s *downloadService) GetPublicManifestPage(ctx context.Context, manifestReq download.PublicManifestRequest) (*download.PublicManifestPage, error) {
	body, err := json.Marshal(manifestReq)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/downloads/public/manifests", s.BaseUrl), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = req.Context()
	}
	page := download.PublicManifestPage{}
	if err := s.client.sendRequest(ctx, req, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

func (s *downloadService) WalkPublicManifest(ctx context.Context, manifestReq download.PublicManifestRequest, fn func(*download.PublicManifestPage) error) error {
	for {
		page, err := s.GetPublicManifestPage(ctx, manifestReq)
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

func (s *downloadService) GetSelection(ctx context.Context, id string) (*download.Selection, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/downloads/selections/%s", s.BaseUrl, url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = req.Context()
	}
	sel := download.Selection{}
	if err := s.client.sendRequest(ctx, req, &sel); err != nil {
		return nil, err
	}
	return &sel, nil
}
