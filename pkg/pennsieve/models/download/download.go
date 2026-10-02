// Package download holds download-service's request and response models.
package download

// ManifestRequest selects files of a dataset for a manifest. Without NodeIds
// the selection is the whole dataset; FileIds narrows NodeIds' packages to
// those files.
type ManifestRequest struct {
	NodeIds []string `json:"nodeIds,omitempty"`
	FileIds []int64  `json:"fileIds,omitempty"`
	// Limit is the files per page, at most (and by default) 1,000.
	Limit int `json:"limit,omitempty"`
	// Cursor continues from a page's Next. WalkManifest sets it.
	Cursor string `json:"cursor,omitempty"`
}

// ManifestPage is one page of a manifest. Header counts the whole selection;
// Data and Blocked are this page's files. Next is empty on the last page.
type ManifestPage struct {
	Header  ManifestHeader `json:"header"`
	Data    []ManifestFile `json:"data"`
	Blocked []BlockedFile  `json:"blocked"`
	Next    string         `json:"next,omitempty"`
}

type ManifestHeader struct {
	Count        int   `json:"count"`
	Size         int64 `json:"size"`
	BlockedCount int   `json:"blockedCount"`
	// ExpiresAt is when this page's first link expires (ISO 8601).
	ExpiresAt string `json:"expiresAt"`
}

// ManifestFile is a file to download: a signed URL and where it goes.
type ManifestFile struct {
	NodeId      string `json:"nodeId"`
	FileId      int64  `json:"fileId"`
	FileName    string `json:"fileName"`
	PackageName string `json:"packageName"`
	// Path is the folder path within the dataset.
	Path       []string `json:"path"`
	URL        string   `json:"url"`
	Size       int64    `json:"size"`
	ScanStatus string   `json:"scanStatus,omitempty"`
}

// BlockedFile is left out because it didn't pass the malware scan.
type BlockedFile struct {
	NodeId      string `json:"nodeId"`
	FileName    string `json:"fileName"`
	PackageName string `json:"packageName"`
	ScanStatus  string `json:"scanStatus"`
}
