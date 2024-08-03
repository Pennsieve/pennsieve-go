package ps_package

// GetPackageSourcesResponse returns from https://api.pennsieve.io/packages/{id}/sources-paged
type GetPackageSourcesResponse struct {
	Limit      int64    `json:"limit"`
	Offset     int64    `json:"offset"`
	Results    []Result `json:"results"`
	TotalCount int64    `json:"totalCount"`
}

type Result struct {
	Content Content `json:"content"`
}

type Content struct {
	PackageID  string   `json:"packageId"`
	Name       string   `json:"name"`
	FileType   string   `json:"fileType"`
	Filename   string   `json:"filename"`
	S3Bucket   string   `json:"s3bucket"`
	S3Key      string   `json:"s3key"`
	ObjectType string   `json:"objectType"`
	Size       int64    `json:"size"`
	CreatedAt  string   `json:"createdAt"`
	UpdatedAt  string   `json:"updatedAt"`
	ID         int64    `json:"id"`
	Checksum   Checksum `json:"checksum"`
}

type Checksum struct {
	ChunkSize int64  `json:"chunkSize"`
	Checksum  string `json:"checksum"`
}

type PresignedFileInfo struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// GetPresignedUrlResponse is returned by GetPresigned URL
type GetPresignedUrlResponse struct {
	Files []PresignedFileInfo `json:"files"`
}

// GetPresignedUrlDTO is returned by the pennsieve API for get presigned URL for sourceFile
type GetPresignedUrlDTO struct {
	URL string `json:"url"`
}
