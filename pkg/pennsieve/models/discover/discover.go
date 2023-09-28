package discover

type GetDatasetByVersionResponse struct {
	ID      int32  `json:"id"`
	Name    string `json:"name"`
	Version int32  `json:"version"`
	Uri     string `json:"uri"`
}

type GetDatasetMetadataByVersionResponse struct {
	ID      int32         `json:"pennsieveDatasetId"`
	Name    string        `json:"name"`
	Version int32         `json:"version"`
	Files   []DatasetFile `json:"files"`
}

type DatasetFile struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	FileType    string `json:"fileType"`
	S3VersionID string `json:"s3VersionId"`
}
