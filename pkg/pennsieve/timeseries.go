package pennsieve

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go/pkg/pennsieve/models/timeseries"
	"net/http"
	"net/url"
	"strconv"
)

type TimeseriesService interface {
	GetChannels(ctx context.Context, datasetId string, packageId string) ([]timeseries.GetChannelsResponse, error)
	GetRangeBlocks(ctx context.Context, datasetId string, packageId string,
		startTime uint64, endTime uint64, channelId string) (*timeseries.GetRangeResponse, error)
}

type timeseriesService struct {
	client  PennsieveHTTPClient
	BaseUrl string
}

func NewTimeseriesService(client PennsieveHTTPClient, baseUrl string) *timeseriesService {
	return &timeseriesService{
		client:  client,
		BaseUrl: baseUrl,
	}
}

func (s *timeseriesService) GetChannels(ctx context.Context, datasetId string, packageId string) ([]timeseries.GetChannelsResponse, error) {

	params := url.Values{}
	params.Add("dataset_id", datasetId)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/timeseries/package/%s/channels?%s", s.BaseUrl, packageId, params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := []timeseries.GetChannelsResponse{}
	if err := s.client.sendRequest(ctx, req, &res); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *timeseriesService) GetRangeBlocks(ctx context.Context, datasetId string,
	packageId string, startTime uint64, endTime uint64, channelId string) (*timeseries.GetRangeResponse, error) {

	params := url.Values{}
	params.Add("dataset_id", datasetId)
	params.Add("start", strconv.FormatUint(startTime, 10))
	params.Add("end", strconv.FormatUint(endTime, 10))

	if len(channelId) != 0 {
		params.Add("channel_id", channelId)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/timeseries/package/%s/range?%s", s.BaseUrl, packageId, params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := timeseries.GetRangeResponse{}
	if err := s.client.sendRequest(ctx, req, &res); err != nil {
		return nil, err
	}

	return &res, nil

}
