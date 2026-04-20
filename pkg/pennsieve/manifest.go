package pennsieve

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest/manifestFile"
)

type ManifestService interface {
	Create(ctx context.Context, requestBody manifest.DTO) (*manifest.PostResponse, error)
	GetFilesForStatus(ctx context.Context, manifestId string,
		status manifestFile.Status, continuationToken string, verify bool) (*manifest.GetStatusEndpointResponse, error)
	GetStorageCredentials(ctx context.Context, datasetId, manifestNodeId string) (*StorageCredentials, error)
	FinalizeManifestFiles(ctx context.Context, datasetId, manifestNodeId string, files []FinalizeFile, opts ...FinalizeOption) (*FinalizeResponse, error)
	SetBaseUrl(url string)
}

// FinalizeOption configures a FinalizeManifestFiles call. Use the provided
// With* helpers (e.g. WithOnConflict) to set values. Unknown options are
// forward-compatible: older SDK versions ignore any options they do not
// recognize, so servers can add new finalize fields without breaking
// existing clients.
type FinalizeOption func(*finalizeOptions)

type finalizeOptions struct {
	onConflict string
}

// FinalizeOnConflict values accepted by the server. Keep in sync with the
// validOnConflict switch in upload-service-v2's finalize_files.go.
const (
	FinalizeOnConflictKeepBoth = "keepBoth"
	FinalizeOnConflictReplace  = "replace"
)

// WithOnConflict tells the server how to resolve name collisions between
// incoming uploads and existing non-deleted packages under the same
// (dataset, folder) tuple. Default (when omitted) is keepBoth — append
// " (N)" to the new upload's name, preserving the existing package.
// Only applies to file packages; folders always get-or-create by name.
func WithOnConflict(v string) FinalizeOption {
	return func(o *finalizeOptions) { o.onConflict = v }
}

// StorageCredentials describes a short-lived STS session scoped to a specific
// manifest's destination prefix. Returned by POST /manifest/storage-credentials.
type StorageCredentials struct {
	AccessKeyID     string    `json:"accessKeyId"`
	SecretAccessKey string    `json:"secretAccessKey"`
	SessionToken    string    `json:"sessionToken"`
	Expiration      time.Time `json:"expiration"`
	Bucket          string    `json:"bucket"`
	KeyPrefix       string    `json:"keyPrefix"`
	Region          string    `json:"region"`
}

// FinalizeFile is one file in a POST /manifest/files/finalize batch.
type FinalizeFile struct {
	UploadID string `json:"uploadId"`
	Size     int64  `json:"size"`
	SHA256   string `json:"sha256,omitempty"`
}

// FinalizeResult is the server's per-file verdict.
type FinalizeResult struct {
	UploadID string `json:"uploadId"`
	Status   string `json:"status"` // "finalized" | "failed"
	Error    string `json:"error,omitempty"`
}

type FinalizeResponse struct {
	Results []FinalizeResult `json:"results"`
}

type manifestService struct {
	client  PennsieveHTTPClient
	baseUrl string
}

func NewManifestService(client PennsieveHTTPClient, baseUrl string) *manifestService {
	return &manifestService{
		client:  client,
		baseUrl: baseUrl,
	}
}

// Create Creates a manifest using the Pensnieve service.
func (s *manifestService) Create(ctx context.Context, requestBody manifest.DTO) (*manifest.PostResponse, error) {

	requestStr := fmt.Sprintf("%s/upload/manifest?dataset_id=%s", s.baseUrl, requestBody.DatasetId)

	body, _ := json.Marshal(requestBody)
	req, err := http.NewRequest("POST", requestStr, bytes.NewBuffer(body))
	if err != nil {
		log.Println("Error: ManifestService.Create || ", err)
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := manifest.PostResponse{}
	if err := s.client.sendRequest(ctx, req, &res); err != nil {

		fmt.Println("SendRequest Error: ", err)
		return nil, err
	}

	return &res, nil
}

// GetFilesForStatus returns a list of files associated with the requested manifest and status.
func (s *manifestService) GetFilesForStatus(ctx context.Context, manifestId string,
	status manifestFile.Status, continuationToken string, verify bool) (*manifest.GetStatusEndpointResponse, error) {

	requestStr := fmt.Sprintf("%s/upload/manifest/status?manifest_id=%s&status=%s&verify=%t", s.baseUrl, manifestId, status, verify)
	if len(continuationToken) > 0 {
		requestStr = requestStr + fmt.Sprintf("&continuation_token=%s", continuationToken)
	}

	req, err := http.NewRequest("GET", requestStr, nil)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = req.Context()
	}

	res := manifest.GetStatusEndpointResponse{}
	if err := s.client.sendRequest(ctx, req, &res); err != nil {
		fmt.Println("SendRequest Error: ", err)
		return nil, err
	}

	return &res, nil

}

func (s *manifestService) SetBaseUrl(url string) {
	s.baseUrl = url
}

// GetStorageCredentials requests STS credentials scoped to the manifest's
// destination storage bucket + O{org}/D{ds}/{manifest}/* prefix. The caller
// can treat HTTPError{StatusCode=404} as "endpoint not yet deployed" and fall
// back to the legacy Cognito + upload-bucket path.
func (s *manifestService) GetStorageCredentials(ctx context.Context, datasetId, manifestNodeId string) (*StorageCredentials, error) {
	requestStr := fmt.Sprintf("%s/upload/manifest/storage-credentials?dataset_id=%s", s.baseUrl, datasetId)

	body, _ := json.Marshal(map[string]string{"manifestNodeId": manifestNodeId})
	req, err := http.NewRequest("POST", requestStr, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = req.Context()
	}

	res := StorageCredentials{}
	if err := s.client.sendRequest(ctx, req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// FinalizeManifestFiles reports a batch of files the agent has successfully
// uploaded direct-to-storage. The server verifies each object, creates
// Postgres package/file rows, and returns per-file results. Idempotent.
// Max batch size is 500 on the server — callers must split larger lists.
//
// Optional FinalizeOptions (e.g. WithOnConflict) tune server-side behavior
// for the batch. Omitting options preserves legacy behavior: name collisions
// auto-rename the new upload to " (N)".
func (s *manifestService) FinalizeManifestFiles(ctx context.Context, datasetId, manifestNodeId string, files []FinalizeFile, opts ...FinalizeOption) (*FinalizeResponse, error) {
	var o finalizeOptions
	for _, fn := range opts {
		fn(&o)
	}

	requestStr := fmt.Sprintf("%s/upload/manifest/files/finalize?dataset_id=%s", s.baseUrl, datasetId)

	body, err := json.Marshal(struct {
		ManifestNodeID string         `json:"manifestNodeId"`
		Files          []FinalizeFile `json:"files"`
		OnConflict     string         `json:"onConflict,omitempty"`
	}{manifestNodeId, files, o.onConflict})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", requestStr, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = req.Context()
	}

	res := FinalizeResponse{}
	if err := s.client.sendRequest(ctx, req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// StorageCredentialsProvider implements aws.CredentialsProvider by fetching
// STS credentials from the Pennsieve storage-credentials endpoint and
// refreshing before they expire. Thread-safe. One provider per
// (manifestNodeId, datasetId) pair.
//
// Concurrency: many S3 workers can call Retrieve simultaneously around the
// refresh boundary. A refresh mutex ensures only one in-flight HTTP refresh;
// others park briefly and pick up the refreshed cache.
//
// Transient errors (network blips, 5xx) are retried with exponential backoff
// so a single upload-service glitch doesn't kill an in-flight upload.
// Permanent errors (4xx) are returned immediately — the caller (agent) can
// fall back to the legacy path.
type StorageCredentialsProvider struct {
	Manifest       ManifestService
	DatasetID      string
	ManifestNodeID string

	mu        sync.Mutex // guards cache read/write
	cache     *StorageCredentials
	refreshMu sync.Mutex // single-flight guard for concurrent refreshes
}

const (
	credsExpiryWindow  = 5 * time.Minute
	credsFetchAttempts = 3
	credsFetchBaseWait = 200 * time.Millisecond
)

// Retrieve returns cached credentials if still valid; otherwise refreshes
// from the API. Safe for concurrent use — only one refresh runs at a time.
func (p *StorageCredentialsProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	// Fast path: valid cached creds.
	if creds, ok := p.cachedIfValid(); ok {
		return creds, nil
	}

	// Slow path: single-flight refresh. While we hold refreshMu, concurrent
	// callers queue behind us; after we finish, they re-check the cache and
	// skip the network call.
	p.refreshMu.Lock()
	defer p.refreshMu.Unlock()

	if creds, ok := p.cachedIfValid(); ok {
		return creds, nil
	}

	c, err := p.fetchWithRetry(ctx)
	if err != nil {
		return aws.Credentials{}, err
	}

	p.mu.Lock()
	p.cache = c
	p.mu.Unlock()

	return toAWSCreds(c), nil
}

// cachedIfValid returns cached credentials if they have more than the expiry
// window left before expiration.
func (p *StorageCredentialsProvider) cachedIfValid() (aws.Credentials, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cache != nil && time.Until(p.cache.Expiration) > credsExpiryWindow {
		return toAWSCreds(p.cache), true
	}
	return aws.Credentials{}, false
}

// fetchWithRetry calls GetStorageCredentials with exponential backoff on
// transient errors. 4xx responses (permanent — auth, not-found, wrong
// dataset) are returned immediately so the caller can fall back.
func (p *StorageCredentialsProvider) fetchWithRetry(ctx context.Context) (*StorageCredentials, error) {
	var lastErr error
	for attempt := 0; attempt < credsFetchAttempts; attempt++ {
		c, err := p.Manifest.GetStorageCredentials(ctx, p.DatasetID, p.ManifestNodeID)
		if err == nil {
			return c, nil
		}
		lastErr = err

		// Don't retry permanent errors.
		var httpErr *HTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode >= 400 && httpErr.StatusCode < 500 {
			return nil, err
		}

		if attempt == credsFetchAttempts-1 {
			break
		}
		delay := credsFetchBaseWait * time.Duration(1<<attempt)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, lastErr
}

// BucketAndPrefix returns the last-fetched bucket and key-prefix, if any.
// Callers should call Retrieve at least once first.
func (p *StorageCredentialsProvider) BucketAndPrefix() (bucket, keyPrefix string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cache == nil {
		return "", ""
	}
	return p.cache.Bucket, p.cache.KeyPrefix
}

// Region returns the last-fetched region. Zero value if no successful fetch yet.
func (p *StorageCredentialsProvider) Region() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cache == nil {
		return ""
	}
	return p.cache.Region
}

func toAWSCreds(c *StorageCredentials) aws.Credentials {
	return aws.Credentials{
		AccessKeyID:     c.AccessKeyID,
		SecretAccessKey: c.SecretAccessKey,
		SessionToken:    c.SessionToken,
		Source:          "PennsieveStorageCredentials",
		CanExpire:       true,
		Expires:         c.Expiration,
	}
}
