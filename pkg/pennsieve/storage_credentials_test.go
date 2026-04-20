package pennsieve

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/manifest/manifestFile"
	"github.com/stretchr/testify/assert"
)

type fakeManifest struct {
	calls       atomic.Int32
	err         error
	transient   int32 // first N calls fail transiently, then succeed
	transientEr error
	creds       StorageCredentials
}

func (f *fakeManifest) Create(ctx context.Context, _ manifest.DTO) (*manifest.PostResponse, error) {
	return nil, nil
}
func (f *fakeManifest) GetFilesForStatus(_ context.Context, _ string, _ manifestFile.Status, _ string, _ bool) (*manifest.GetStatusEndpointResponse, error) {
	return nil, nil
}
func (f *fakeManifest) SetBaseUrl(_ string) {}
func (f *fakeManifest) FinalizeManifestFiles(_ context.Context, _, _ string, _ []FinalizeFile, _ ...FinalizeOption) (*FinalizeResponse, error) {
	return nil, nil
}
func (f *fakeManifest) GetStorageCredentials(_ context.Context, _, _ string) (*StorageCredentials, error) {
	n := f.calls.Add(1)
	if n <= f.transient && f.transientEr != nil {
		return nil, f.transientEr
	}
	if f.err != nil {
		return nil, f.err
	}
	c := f.creds
	return &c, nil
}

func TestStorageCredsProvider_CachesValidCreds(t *testing.T) {
	fm := &fakeManifest{creds: StorageCredentials{
		AccessKeyID: "AK", SecretAccessKey: "SK", SessionToken: "ST",
		Expiration: time.Now().Add(1 * time.Hour),
		Bucket:     "b", KeyPrefix: "O1/D2/m", Region: "us-east-1",
	}}
	p := &StorageCredentialsProvider{Manifest: fm}

	for i := 0; i < 5; i++ {
		_, err := p.Retrieve(context.Background())
		assert.NoError(t, err)
	}
	assert.Equal(t, int32(1), fm.calls.Load(), "should cache — only first call fetches")
}

func TestStorageCredsProvider_RefreshesNearExpiry(t *testing.T) {
	fm := &fakeManifest{creds: StorageCredentials{
		Expiration: time.Now().Add(2 * time.Minute), // within 5-min window
		Region:     "us-east-1",
	}}
	p := &StorageCredentialsProvider{Manifest: fm}

	_, _ = p.Retrieve(context.Background())
	_, _ = p.Retrieve(context.Background())
	assert.Equal(t, int32(2), fm.calls.Load(), "should refetch because cached creds are within expiry window")
}

func TestStorageCredsProvider_SingleFlight(t *testing.T) {
	// Block the fake so we can observe multiple goroutines piling up.
	release := make(chan struct{})
	var started atomic.Int32
	fm := &fakeManifest{creds: StorageCredentials{
		Expiration: time.Now().Add(1 * time.Hour),
		Region:     "us-east-1",
	}}
	// Wrap to add a block.
	blocking := &blockingManifest{inner: fm, started: &started, release: release}
	p := &StorageCredentialsProvider{Manifest: blocking}

	const n = 10
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = p.Retrieve(context.Background())
		}()
	}

	// Wait until the first Retrieve has started its network call.
	for started.Load() == 0 {
		time.Sleep(time.Millisecond)
	}
	close(release)
	wg.Wait()

	assert.Equal(t, int32(1), fm.calls.Load(), "single-flight should collapse concurrent refreshes to one")
}

type blockingManifest struct {
	inner   *fakeManifest
	started *atomic.Int32
	release chan struct{}
}

func (b *blockingManifest) Create(ctx context.Context, d manifest.DTO) (*manifest.PostResponse, error) {
	return b.inner.Create(ctx, d)
}
func (b *blockingManifest) GetFilesForStatus(ctx context.Context, id string, s manifestFile.Status, ct string, v bool) (*manifest.GetStatusEndpointResponse, error) {
	return b.inner.GetFilesForStatus(ctx, id, s, ct, v)
}
func (b *blockingManifest) SetBaseUrl(u string) { b.inner.SetBaseUrl(u) }
func (b *blockingManifest) FinalizeManifestFiles(ctx context.Context, d, m string, f []FinalizeFile, opts ...FinalizeOption) (*FinalizeResponse, error) {
	return b.inner.FinalizeManifestFiles(ctx, d, m, f, opts...)
}
func (b *blockingManifest) GetStorageCredentials(ctx context.Context, d, m string) (*StorageCredentials, error) {
	b.started.Add(1)
	<-b.release
	return b.inner.GetStorageCredentials(ctx, d, m)
}

func TestStorageCredsProvider_RetriesTransient(t *testing.T) {
	fm := &fakeManifest{
		transient:   2, // first 2 calls fail with 5xx
		transientEr: &HTTPError{StatusCode: 503, Message: "unavailable"},
		creds: StorageCredentials{
			Expiration: time.Now().Add(1 * time.Hour),
			Region:     "us-east-1",
		},
	}
	p := &StorageCredentialsProvider{Manifest: fm}

	_, err := p.Retrieve(context.Background())
	assert.NoError(t, err, "should succeed on 3rd attempt")
	assert.Equal(t, int32(3), fm.calls.Load())
}

func TestStorageCredsProvider_DoesNotRetry4xx(t *testing.T) {
	fm := &fakeManifest{
		transient:   5,
		transientEr: &HTTPError{StatusCode: 404, Message: "not found"},
	}
	p := &StorageCredentialsProvider{Manifest: fm}

	_, err := p.Retrieve(context.Background())
	assert.Error(t, err)
	var httpErr *HTTPError
	assert.True(t, errors.As(err, &httpErr))
	assert.Equal(t, 404, httpErr.StatusCode)
	assert.Equal(t, int32(1), fm.calls.Load(), "permanent errors should not be retried")
}
