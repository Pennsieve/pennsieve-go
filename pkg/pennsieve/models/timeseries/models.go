package timeseries

type GetChannelsResponse struct {
	ChannelID string  `json:"channelId"`
	Name      string  `json:"name"`
	StartTime int     `json:"startTime"`
	EndTime   int     `json:"endTime"`
	Unit      string  `json:"unit"`
	Rate      float64 `json:"rate"`
}

// GetRangeResponse - RESPONSE FROM GET RANGE REQUEST
type GetRangeResponse struct {
	RequestedStartTime int64      `json:"requestedStartTime"`
	RequestedEndTime   int64      `json:"requestedEndTime"`
	Channels           []Channels `json:"channels"`
}
type Ranges struct {
	ID           string `json:"id"`
	StartTime    int64  `json:"startTime"`
	EndTime      int64  `json:"endTime"`
	SamplingRate int    `json:"samplingRate"`
	PreSignedURL string `json:"preSignedUrl"`
}
type Channels struct {
	ChannelID string   `json:"channelId"`
	Name      string   `json:"name"`
	StartTime uint64   `json:"startTime"`
	EndTime   uint64   `json:"endTime"`
	Unit      string   `json:"unit"`
	Rate      float64  `json:"rate"`
	Ranges    []Ranges `json:"ranges"`
}
