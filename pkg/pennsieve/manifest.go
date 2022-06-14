package pennsieve

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pennsieve/pennsieve-go-api/models/manifest"
	"net/http"
)

type ManifestResonse struct {
	ManifestNodeId string   `json:"manifest_node_id"'`
	NrFilesAdded   int      `json:"nrFilesAdded"`
	FailedFiles    []string `json:"failedFiles"`
}

type ManifestService struct {
	client  *Client
	baseUrl string
}

// Get returns a single dataset by id.
func (d *ManifestService) Create(ctx context.Context, requestBody manifest.DTO) (*ManifestResonse, error) {

	fmt.Println("baseurl for manifestService: ", d.baseUrl)

	requestStr := fmt.Sprintf("%s/manifest?dataset_id=%s", d.baseUrl, requestBody.DatasetId)

	fmt.Println("Request: ", requestStr)
	body, _ := json.Marshal(requestBody)
	req, err := http.NewRequest("POST", requestStr, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := ManifestResonse{}
	if err := d.client.sendRequest(ctx, req, &res); err != nil {

		fmt.Println("SendRequest Error: ", err)
		return nil, err
	}

	return &res, nil
}
