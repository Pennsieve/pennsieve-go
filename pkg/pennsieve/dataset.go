package pennsieve

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/dataset"
)

type DatasetService interface {
	Get(ctx context.Context, id string) (*dataset.GetDatasetResponse, error)
	Find(ctx context.Context, limit int, query string) (*dataset.ListDatasetResponse, error)
	List(ctx context.Context, limit int, offset int) (*dataset.ListDatasetResponse, error)
	SetBaseUrl(url string)
	Create(ctx context.Context, name, description, tags string) (*dataset.CreateDatasetResponse, error)
	GetDatasetByVersion(ctx context.Context, datasetId int32, versionId int32) (*dataset.GetDatasetByVersionResponse, error)
	GetDatasetMetadataByVersion(ctx context.Context, datasetId int32, versionId int32) (*dataset.GetDatasetMetadataByVersionResponse, error)
}

type datasetService struct {
	Client  PennsieveHTTPClient
	BaseUrl string
}

func NewDatasetService(client PennsieveHTTPClient, baseUrl string) *datasetService {
	return &datasetService{
		Client:  client,
		BaseUrl: baseUrl,
	}
}

// Get returns a single dataset by id.
func (d *datasetService) Get(ctx context.Context, id string) (*dataset.GetDatasetResponse, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/datasets/%s", d.BaseUrl, id), nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := dataset.GetDatasetResponse{}
	if err := d.Client.sendRequest(ctx, req, &res); err != nil {

		log.Println("DatasetService: SendRequest Error in Get: ", err)
		return nil, err
	}

	return &res, nil
}

// Find returns a list of datasets based on a provided query
func (d *datasetService) Find(ctx context.Context, limit int, query string) (*dataset.ListDatasetResponse, error) {

	params := url.Values{}
	params.Add("limit", strconv.Itoa(limit))
	params.Add("query", query)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/datasets/paginated?%s", d.BaseUrl, params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := dataset.ListDatasetResponse{}
	if err := d.Client.sendRequest(ctx, req, &res); err != nil {

		fmt.Println("SendRequest Error: ", err)
		return nil, err
	}

	return &res, nil
}

// List returns a list of datasets with a limit and an offset
func (d *datasetService) List(ctx context.Context, limit int, offset int) (*dataset.ListDatasetResponse, error) {

	params := url.Values{}
	params.Add("limit", strconv.Itoa(limit))
	params.Add("offset", strconv.Itoa(offset))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/datasets/paginated?%s", d.BaseUrl, params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := dataset.ListDatasetResponse{}
	if err := d.Client.sendRequest(ctx, req, &res); err != nil {

		fmt.Println("SendRequest Error: ", err)
		return nil, err
	}

	return &res, nil
}

func (s *datasetService) SetBaseUrl(url string) {
	s.BaseUrl = url
}

func (d *datasetService) Create(ctx context.Context, name, description, tags string) (*dataset.CreateDatasetResponse, error) {
	postParams := fmt.Sprintf(`
		{
			"name": "%s",
			"description": "%s",
			"tags": %v
		}`, name, description, tags)

	postParamsPayload := bytes.NewReader([]byte(postParams))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/datasets/", d.BaseUrl), postParamsPayload)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := dataset.CreateDatasetResponse{}
	if err := d.Client.sendRequest(ctx, req, &res); err != nil {
		fmt.Println("SendRequest Error: ", err)
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return &res, nil

}

// GetDatasetByVersion returns a dataset by version.
func (d *datasetService) GetDatasetByVersion(ctx context.Context, datasetId int32, versionId int32) (*dataset.GetDatasetByVersionResponse, error) {
	endpoint := fmt.Sprintf("%s/discover/datasets/%v/versions/%v",
		d.BaseUrl, datasetId, versionId)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := dataset.GetDatasetByVersionResponse{}
	if err := d.Client.sendUnauthenticatedRequest(ctx, req, &res); err != nil {
		log.Println("DatasetService: SendRequest Error in GetDatasetByVersion: ", err)
		return nil, err
	}

	return &res, nil
}

// GetDatasetMetadataByVersion returns dataset metadata by version.
func (d *datasetService) GetDatasetMetadataByVersion(ctx context.Context, datasetId int32, versionId int32) (*dataset.GetDatasetMetadataByVersionResponse, error) {
	endpoint := fmt.Sprintf("%s/discover/datasets/%v/versions/%v/metadata",
		d.BaseUrl, datasetId, versionId)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := dataset.GetDatasetMetadataByVersionResponse{}
	if err := d.Client.sendUnauthenticatedRequest(ctx, req, &res); err != nil {
		log.Println("DatasetService: SendRequest Error in GetDatasetMetadataByVersion: ", err)
		return nil, err
	}

	return &res, nil
}
