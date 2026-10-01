package services

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoogleSheetExportURL(t *testing.T) {
	got, err := googleSheetExportURL("https://docs.google.com/spreadsheets/d/1SViTcU7Tp6QAV1ExampleSheetID/edit?gid=123#gid=123")
	require.NoError(t, err)
	require.Equal(t, "https://docs.google.com/spreadsheets/d/1SViTcU7Tp6QAV1ExampleSheetID/export?format=xlsx", got)
}

func TestGoogleSheetExportURLRejectsNonGoogleHost(t *testing.T) {
	_, err := googleSheetExportURL("https://example.com/spreadsheets/d/1SViTcU7Tp6QAV1ExampleSheetID/edit")
	require.Error(t, err)
}
