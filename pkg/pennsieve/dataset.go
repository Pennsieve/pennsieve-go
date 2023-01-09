package pennsieve

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/dataset"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type DatasetService interface {
	Get(ctx context.Context, id string) (*dataset.GetDatasetResponse, error)
	Find(ctx context.Context, limit int, query string) (*dataset.ListDatasetResponse, error)
	List(ctx context.Context, limit int, offset int) (*dataset.ListDatasetResponse, error)
	SetBaseUrl(url string)
	Create(ctx context.Context, name, description string, tags []string)
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

func (d *datasetService) Create(ctx context.Context, name, description string, tags []string) (*dataset.CreateDatasetResponse, error) {

	params := url.Values{}

	params.Add("name", name)
	params.Add("description", description)

	for i := 0; i < len(tags); i++ {
		params.Add("tags", tags[i])
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/datasets/", d.BaseUrl), strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
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
