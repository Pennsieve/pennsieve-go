package pennsieve

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetDataset(t *testing.T) {

	ht := mockHTTPClient{
		APISession:         APISession{},
		APICredentials:     APICredentials{},
		OrganizationNodeId: "",
		OrganizationId:     0,
	}

	testDatasetService := NewDatasetService(&ht, "https://test.com")
	dataset, _ := testDatasetService.Get(context.Background(), "N:Dataset:1234")
	assert.Equal(t, "Test Dataset", dataset.Content.Name, "Org name must match.")
}
