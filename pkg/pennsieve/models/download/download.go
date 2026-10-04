// Package download holds download-service's request and response models.
package download

// ManifestRequest selects files of a dataset for a manifest: packages and
// folders (a package is one file), or a saved selection. Without either the
// selection is the whole dataset.
type ManifestRequest struct {
	NodeIds []string `json:"nodeIds,omitempty"`
	// SelectionId downloads a saved selection of the dataset, in place of
	// NodeIds.
	SelectionId string `json:"selectionId,omitempty"`
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

// PublicManifestRequest selects files of a published (Discover) dataset
// version: paths, or a saved public selection.
type PublicManifestRequest struct {
	DatasetId int64 `json:"datasetId,omitempty"`
	// Version is a version number; zero is the latest.
	Version int `json:"version,omitempty"`
	// Paths are files and folders of the version; none is the whole version.
	Paths []string `json:"paths,omitempty"`
	// SelectionId downloads a saved public selection, in place of DatasetId,
	// Version and Paths.
	SelectionId string `json:"selectionId,omitempty"`
	// Limit is the files per page, at most (and by default) 1,000.
	Limit int `json:"limit,omitempty"`
	// Cursor continues from a page's Next. WalkPublicManifest sets it.
	Cursor string `json:"cursor,omitempty"`
}

// PublicManifestPage is one page of a published selection. Header counts the
// whole selection; Data and Skipped are this page's files.
type PublicManifestPage struct {
	Header  PublicManifestHeader `json:"header"`
	Data    []PublicManifestFile `json:"data"`
	Skipped []SkippedFile        `json:"skipped"`
	Next    string               `json:"next,omitempty"`
}

type PublicManifestHeader struct {
	DatasetId int64 `json:"datasetId"`
	// Version is the version resolved, also when "latest" was asked for.
	Version int   `json:"version"`
	Count   int   `json:"count"`
	Size    int64 `json:"size"`
	// ExpiresAt is when this page's first link expires (ISO 8601).
	ExpiresAt string `json:"expiresAt"`
}

// PublicManifestFile is a published file to download. SHA256, when present,
// is the checksum recorded at publication.
type PublicManifestFile struct {
	FileName string   `json:"fileName"`
	Path     []string `json:"path"`
	URL      string   `json:"url"`
	Size     int64    `json:"size"`
	SHA256   string   `json:"sha256,omitempty"`
}

// SkippedFile is a published file that can't be downloaded, and why.
type SkippedFile struct {
	Path     []string `json:"path"`
	FileName string   `json:"fileName"`
	Reason   string   `json:"reason"`
}

// SelectionKind says which manifest a saved selection is downloaded with.
type SelectionKind string

const (
	// SelectionWorkspace: WalkManifest with the selection's DatasetNodeId.
	SelectionWorkspace SelectionKind = "workspace"
	// SelectionPublic: WalkPublicManifest.
	SelectionPublic SelectionKind = "public"
)

// Selection says where to download a saved selection from. It holds no file
// details: the manifest, requested with its id, resolves the files with the
// downloader's own access.
type Selection struct {
	Id   string        `json:"id"`
	Kind SelectionKind `json:"kind"`
	// DatasetNodeId is a workspace selection's dataset.
	DatasetNodeId string `json:"datasetNodeId,omitempty"`
	// DatasetId and Version are a public selection's published version.
	DatasetId int64  `json:"datasetId,omitempty"`
	Version   int    `json:"version,omitempty"`
	ExpiresAt string `json:"expiresAt"`
}
