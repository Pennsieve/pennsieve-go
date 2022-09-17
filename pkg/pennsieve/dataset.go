package pennsieve

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models"
	"net/http"
)

type DatasetService struct {
	Client  *Client
	BaseUrl string
}

// Get returns a single dataset by id.
func (d *DatasetService) Get(ctx context.Context, id string) (*models.GetDatasetResponse, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/datasets/%s", d.BaseUrl, id), nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := models.GetDatasetResponse{}
	if err := d.Client.SendRequest(ctx, req, &res); err != nil {

		fmt.Println("SendRequest Error: ", err)
		return nil, err
	}

	return &res, nil
}

func (d *DatasetService) List(ctx context.Context, limit int, offset int) (*models.ListDatasetResponse, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/datasets/paginated?limit=%d&offset=%d", d.BaseUrl, limit, offset), nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := models.ListDatasetResponse{}
	if err := d.Client.SendRequest(ctx, req, &res); err != nil {

		fmt.Println("SendRequest Error: ", err)
		return nil, err
	}

	return &res, nil
}
