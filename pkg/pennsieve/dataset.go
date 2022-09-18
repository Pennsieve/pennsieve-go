package pennsieve

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/dataset"
	"net/http"
	"net/url"
	"strconv"
)

type DatasetService struct {
	Client  *Client
	BaseUrl string
}

// Get returns a single dataset by id.
func (d *DatasetService) Get(ctx context.Context, id string) (*dataset.GetDatasetResponse, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/datasets/%s", d.BaseUrl, id), nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := dataset.GetDatasetResponse{}
	if err := d.Client.SendRequest(ctx, req, &res); err != nil {

		fmt.Println("SendRequest Error: ", err)
		return nil, err
	}

	return &res, nil
}

// Find returns a list of datasets based on a provided query
func (d *DatasetService) Find(ctx context.Context, limit int, query string) (*dataset.ListDatasetResponse, error) {

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
	if err := d.Client.SendRequest(ctx, req, &res); err != nil {

		fmt.Println("SendRequest Error: ", err)
		return nil, err
	}

	return &res, nil
}

// List returns a list of datasets with a limit and an offset
func (d *DatasetService) List(ctx context.Context, limit int, offset int) (*dataset.ListDatasetResponse, error) {

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
	if err := d.Client.SendRequest(ctx, req, &res); err != nil {

		fmt.Println("SendRequest Error: ", err)
		return nil, err
	}

	return &res, nil
}
