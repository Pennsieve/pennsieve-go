package pennsieve

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/discover"
)

type DiscoverService interface {
	GetDatasetByVersion(ctx context.Context, datasetId int32, versionId int32) (*discover.GetDatasetByVersionResponse, error)
	GetDatasetMetadataByVersion(ctx context.Context, datasetId int32, versionId int32) (*discover.GetDatasetMetadataByVersionResponse, error)
	GetDatasetFileByVersion(ctx context.Context, datasetId int32, versionId int32, filename string) (*discover.GetDatasetFileByVersionResponse, error)
}

type discoverService struct {
	Client  PennsieveHTTPClient
	BaseUrl string
}

func NewDiscoverService(client PennsieveHTTPClient, baseUrl string) *discoverService {
	return &discoverService{
		Client:  client,
		BaseUrl: baseUrl,
	}
}

// GetDatasetByVersion returns a dataset by version.
func (d *discoverService) GetDatasetByVersion(ctx context.Context, datasetId int32, versionId int32) (*discover.GetDatasetByVersionResponse, error) {
	endpoint := fmt.Sprintf("%s/discover/datasets/%v/versions/%v",
		d.BaseUrl, datasetId, versionId)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := discover.GetDatasetByVersionResponse{}
	if err := d.Client.sendUnauthenticatedRequest(ctx, req, &res); err != nil {
		log.Println("DatasetService: SendRequest Error in GetDatasetByVersion: ", err)
		return nil, err
	}

	return &res, nil
}

// GetDatasetMetadataByVersion returns dataset metadata by version.
func (d *discoverService) GetDatasetMetadataByVersion(ctx context.Context, datasetId int32, versionId int32) (*discover.GetDatasetMetadataByVersionResponse, error) {
	endpoint := fmt.Sprintf("%s/discover/datasets/%v/versions/%v/metadata",
		d.BaseUrl, datasetId, versionId)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := discover.GetDatasetMetadataByVersionResponse{}
	if err := d.Client.sendUnauthenticatedRequest(ctx, req, &res); err != nil {
		log.Println("DatasetService: SendRequest Error in GetDatasetMetadataByVersion: ", err)
		return nil, err
	}

	return &res, nil
}

// GetDatasetFileByVersion returns a specific dataset file by version.
func (d *discoverService) GetDatasetFileByVersion(ctx context.Context, datasetId int32, versionId int32, filename string) (*discover.GetDatasetFileByVersionResponse, error) {
	endpoint := fmt.Sprintf("%s/discover/datasets/%v/versions/%v/files?path=%v",
		d.BaseUrl, datasetId, versionId, filename)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := discover.GetDatasetFileByVersionResponse{}
	if err := d.Client.sendUnauthenticatedRequest(ctx, req, &res); err != nil {
		log.Println("DatasetService: SendRequest Error in GetDatasetFileByVersion: ", err)
		return nil, err
	}

	return &res, nil
}
