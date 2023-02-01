package pennsieve

import (
	"context"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateManifest(t *testing.T) {

	client := NewClient(APIParams{}, "")

	testManifestService := client.Manifest
	manifest, _ := testManifestService.Create(context.Background(), manifest.DTO{
		ID:        "",
		DatasetId: "",
		Files:     nil,
		Status:    0,
	})
	assert.Equal(t, "1234", manifest.ManifestNodeId, "Manifest ID name must match.")
}
