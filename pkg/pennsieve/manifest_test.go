package pennsieve

import (
	"context"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateManifest(t *testing.T) {

	ht := mockHTTPClient{
		APISession:         APISession{},
		APICredentials:     APICredentials{},
		OrganizationNodeId: "",
		OrganizationId:     0,
	}

	testManifestService := NewManifestService(&ht, "https://test.com")
	manifest, _ := testManifestService.Create(context.Background(), manifest.DTO{
		ID:        "",
		DatasetId: "",
		Files:     nil,
		Status:    0,
	})
	assert.Equal(t, "1234", manifest.ManifestNodeId, "Manifest ID name must match.")
}
