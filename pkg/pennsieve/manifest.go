package pennsieve

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest/manifestFile"
	"log"
	"net/http"
)

type ManifestService interface {
	Create(ctx context.Context, requestBody manifest.DTO) (*manifest.PostResponse, error)
	GetFilesForStatus(ctx context.Context, manifestId string,
		status manifestFile.Status, continuationToken string, verify bool) (*manifest.GetStatusEndpointResponse, error)
	SetBaseUrl(url string)
}

type manifestService struct {
	client  PennsieveHTTPClient
	baseUrl string
}

func NewManifestService(client PennsieveHTTPClient, baseUrl string) *manifestService {
	return &manifestService{
		client:  client,
		baseUrl: baseUrl,
	}
}

// Create Creates a manifest using the Pensnieve service.
func (s *manifestService) Create(ctx context.Context, requestBody manifest.DTO) (*manifest.PostResponse, error) {

	requestStr := fmt.Sprintf("%s/manifest?dataset_id=%s", s.baseUrl, requestBody.DatasetId)

	body, _ := json.Marshal(requestBody)
	req, err := http.NewRequest("POST", requestStr, bytes.NewBuffer(body))
	if err != nil {
		log.Println("Error: ManifestService.Create || ", err)
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := manifest.PostResponse{}
	if err := s.client.sendRequest(ctx, req, &res); err != nil {

		fmt.Println("SendRequest Error: ", err)
		return nil, err
	}

	return &res, nil
}

// GetFilesForStatus returns a list of files associated with the requested manifest and status.
func (s *manifestService) GetFilesForStatus(ctx context.Context, manifestId string,
	status manifestFile.Status, continuationToken string, verify bool) (*manifest.GetStatusEndpointResponse, error) {

	requestStr := fmt.Sprintf("%s/manifest/status?manifest_id=%s&status=%s&verify=%t", s.baseUrl, manifestId, status, verify)
	if len(continuationToken) > 0 {
		requestStr = requestStr + fmt.Sprintf("&continuation_token=%s", continuationToken)
	}

	req, err := http.NewRequest("GET", requestStr, nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := manifest.GetStatusEndpointResponse{}
	if err := s.client.sendRequest(ctx, req, &res); err != nil {
		fmt.Println("SendRequest Error: ", err)
		return nil, err
	}

	return &res, nil

}

func (s *manifestService) SetBaseUrl(url string) {
	s.baseUrl = url
}
