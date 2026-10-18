package download

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifestPageDecodesPendingFiles(t *testing.T) {
	body := `{
		"header": {"count": 1, "size": 4, "blockedCount": 1, "pendingCount": 3, "expiresAt": "2026-10-17T12:00:00Z"},
		"data": [{"nodeId": "N:package:1", "fileId": 1, "fileName": "a.csv", "path": [], "url": "https://s3/a", "size": 4}],
		"blocked": [{"nodeId": "N:package:9", "fileName": "virus.exe", "packageName": "virus", "scanStatus": "infected"}],
		"pending": [{"nodeId": "N:package:7", "fileName": "new.edf", "packageName": "new", "scanStatus": "scanning"}]
	}`
	var page ManifestPage
	require.NoError(t, json.Unmarshal([]byte(body), &page))
	assert.Equal(t, 3, page.Header.PendingCount, "the header counts the whole selection")
	assert.Equal(t, []PendingFile{{NodeId: "N:package:7", FileName: "new.edf", PackageName: "new", ScanStatus: "scanning"}}, page.Pending)
	assert.Len(t, page.Data, 1)
	assert.Len(t, page.Blocked, 1)
}

func TestManifestPageWithoutPendingFiles(t *testing.T) {
	// download-service without the scan wait gate sends no pending files.
	var page ManifestPage
	require.NoError(t, json.Unmarshal([]byte(`{"header": {"count": 0}, "data": [], "blocked": []}`), &page))
	assert.Zero(t, page.Header.PendingCount)
	assert.Empty(t, page.Pending)
}
