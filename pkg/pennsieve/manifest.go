package pennsieve

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pennsieve/pennsieve-go-api/models/manifest"
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
	if err := d.client.sendRequest(ctx, req, &res); err != nil {

		fmt.Println("SendRequest Error: ", err)
		return nil, err
	}

	return &res, nil
}
