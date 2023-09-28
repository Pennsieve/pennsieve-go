package dataset

import (
	"time"
)

type GetDatasetResponse struct {
	Content            Content            `json:"content"`
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
type Content struct {
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

type ListDatasetResponse struct {
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
	TotalCount int        `json:"totalCount"`
	Datasets   []Datasets `json:"datasets"`
}

type CreateDatasetResponse struct {
	Content      Content       `json:"content"`
	Organization string        `json:"organization"`
	Publication  Publication   `json:"publication,omitempty"`
	Properties   []interface{} `json:"properties"`
	CanPublish   bool          `json:"canPublish"`
	Locked       bool          `json:"locked"`
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
	Status             string `json:"status"`
	Type               string `json:"type,omitempty"`
	EmbargoReleaseDate string `json:"embargoReleaseDate,omitempty"`
}
type Datasets struct {
	Content            Content            `json:"content"`
	Organization       string             `json:"organization"`
	Owner              string             `json:"owner"`
	CollaboratorCounts CollaboratorCounts `json:"collaboratorCounts"`
	Storage            int                `json:"storage"`
	Status             Status             `json:"status"`
	Properties         []interface{}      `json:"properties"`
	CanPublish         bool               `json:"canPublish"`
	Locked             bool               `json:"locked"`
	Publication        Publication        `json:"publication,omitempty"`
}
