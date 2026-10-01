package services

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAmaliyahWorkbookAcceptsRenamedSingleSheet(t *testing.T) {
	original := buildTrackerAmaliyahTestWorkbook(t)
	reader, err := zip.NewReader(bytes.NewReader(original), int64(len(original)))
	require.NoError(t, err)

	var out bytes.Buffer
	writer := zip.NewWriter(&out)
	for _, file := range reader.File {
		r, err := file.Open()
		require.NoError(t, err)
		data, err := io.ReadAll(r)
		require.NoError(t, err)
		require.NoError(t, r.Close())

		if file.Name == "xl/workbook.xml" {
			data = []byte(strings.ReplaceAll(string(data), "Tracker Amaliyah Harian", "Workspace Indhy"))
		}

		w, err := writer.Create(file.Name)
		require.NoError(t, err)
		_, err = w.Write(data)
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())

	metrics, err := parseAmaliyahWorkbook(out.Bytes(), "2026-09")
	require.NoError(t, err)
	require.Len(t, metrics, 4)
}
