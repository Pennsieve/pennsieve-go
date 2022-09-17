package pennsieve

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/manifest"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/manifest/manifestFile"
	"net/http"
)

type ManifestService struct {
	client  *Client
	baseUrl string
}

// Get returns a single dataset by id.
func (d *ManifestService) Create(ctx context.Context, requestBody manifest.DTO) (*manifest.PostResponse, error) {

	requestStr := fmt.Sprintf("%s/manifest?dataset_id=%s", d.baseUrl, requestBody.DatasetId)

	body, _ := json.Marshal(requestBody)
	req, err := http.NewRequest("POST", requestStr, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := manifest.PostResponse{}
	if err := d.client.SendRequest(ctx, req, &res); err != nil {

		fmt.Println("SendRequest Error: ", err)
		return nil, err
	}

	return &res, nil
}

// GetFilesForStatus returns a list of files associated with the requested manifest and status.
func (d *ManifestService) GetFilesForStatus(ctx context.Context, manifestId string,
	status manifestFile.Status, continuationToken string, verify bool) (*manifest.GetStatusEndpointResponse, error) {

	requestStr := fmt.Sprintf("%s/manifest/status?manifest_id=%s&status=%s&verify=%t", d.baseUrl, manifestId, status, verify)
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
	if err := d.client.SendRequest(ctx, req, &res); err != nil {

		fmt.Println("SendRequest Error: ", err)
		return nil, err
	}

	return &res, nil

}
