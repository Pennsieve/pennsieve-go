package pennsieve

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type GetDatasetResponse struct {
	Content            DatasetContent     `json:"content"`
	Organization       string             `json:"organization"`
	Children           []Children         `json:"children"`
	Owner              string             `json:"owner"`
	CollaboratorCounts CollaboratorCounts `json:"collaboratorCounts"`
	Storage            int                `json:"storage"`
	Status             Status             `json:"status"`
	Publication        Publication        `json:"publication"`
	Properties         []interface{}      `json:"properties"`
	CanPublish         bool               `json:"canPublish"`
	Locked             bool               `json:"locked"`
	BannerPresignedURL string             `json:"bannerPresignedUrl"`
}
type DatasetContent struct {
	ID                           string    `json:"id"`
	Name                         string    `json:"name"`
	Description                  string    `json:"description"`
	State                        string    `json:"state"`
	CreatedAt                    time.Time `json:"createdAt"`
	UpdatedAt                    time.Time `json:"updatedAt"`
	PackageType                  string    `json:"packageType"`
	DatasetType                  string    `json:"datasetType"`
	Status                       string    `json:"status"`
	AutomaticallyProcessPackages bool      `json:"automaticallyProcessPackages"`
	License                      string    `json:"license"`
	Tags                         []string  `json:"tags"`
	IntID                        int       `json:"intId"`
}
type ChildrenContent struct {
	ID            string    `json:"id"`
	NodeID        string    `json:"nodeId"`
	Name          string    `json:"name"`
	PackageType   string    `json:"packageType"`
	DatasetID     string    `json:"datasetId"`
	DatasetNodeID string    `json:"datasetNodeId"`
	OwnerID       int       `json:"ownerId"`
	State         string    `json:"state"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	IntID         int       `json:"intId"`
	DatasetIntID  int       `json:"datasetIntId"`
}
type Properties struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	DataType string `json:"dataType"`
	Fixed    bool   `json:"fixed"`
	Hidden   bool   `json:"hidden"`
	Display  string `json:"display"`
}
type PropsWithCategory struct {
	Category   string       `json:"category"`
	Properties []Properties `json:"properties"`
}
type Children struct {
	Content    ChildrenContent     `json:"content"`
	Properties []PropsWithCategory `json:"properties"`
	Children   []interface{}       `json:"children"`
	Storage    int                 `json:"storage"`
	Extension  string              `json:"extension,omitempty"`
}
type CollaboratorCounts struct {
	Users         int `json:"users"`
	Organizations int `json:"organizations"`
	Teams         int `json:"teams"`
}
type Status struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Color       string `json:"color"`
	InUse       bool   `json:"inUse"`
}
type Publication struct {
	Status string `json:"status"`
	Type   string `json:"type"`
}

type DatasetService struct {
	client  *Client
	baseUrl string
}

// Get returns a single dataset by id.
func (d *DatasetService) Get(ctx context.Context, id string) (*GetDatasetResponse, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/datasets/%s", d.baseUrl, id), nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := GetDatasetResponse{}
	if err := d.client.sendRequest(ctx, req, &res); err != nil {

		fmt.Println("SendRequest Error: ", err)
		return nil, err
	}

	return &res, nil
}
