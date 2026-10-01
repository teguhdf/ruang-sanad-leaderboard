package services

import (
	"archive/zip"
	"bytes"
	"fmt"
	"testing"

	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/stretchr/testify/require"
)

func TestParseAmaliyahWorkbookDynamicCheckboxDenominator(t *testing.T) {
	data := buildAmaliyahTestWorkbook(t)

	metrics, err := parseAmaliyahWorkbook(data, "2026-09")
	require.NoError(t, err)
	require.Len(t, metrics, 4)

	byKey := make(map[string]models.AmaliyahMetricScore, len(metrics))
	for _, metric := range metrics {
		byKey[metric.Key] = metric
	}

	require.Equal(t, 3, byKey[models.AmaliyahTahajud].Available)
	require.Equal(t, 2, byKey[models.AmaliyahTahajud].Completed)
	require.InDelta(t, 66.6667, byKey[models.AmaliyahTahajud].Score, 0.001)

	require.Equal(t, 3, byKey[models.AmaliyahTilawahTadabur].Available)
	require.Equal(t, 2, byKey[models.AmaliyahTilawahTadabur].Completed)
	require.InDelta(t, 66.6667, byKey[models.AmaliyahTilawahTadabur].Score, 0.001)

	require.Equal(t, 4, byKey[models.AmaliyahDzikirPagi].Available)
	require.Equal(t, 2, byKey[models.AmaliyahDzikirPagi].Completed)
	require.Equal(t, float64(50), byKey[models.AmaliyahDzikirPagi].Score)

	require.Equal(t, 4, byKey[models.AmaliyahDzikirPetang].Available)
	require.Zero(t, byKey[models.AmaliyahDzikirPetang].Completed)
	require.Zero(t, byKey[models.AmaliyahDzikirPetang].Score)
}

func TestParseAmaliyahWorkbookTrackerLayout(t *testing.T) {
	data := buildTrackerAmaliyahTestWorkbook(t)

	metrics, err := parseAmaliyahWorkbook(data, "2026-09")
	require.NoError(t, err)
	require.Len(t, metrics, 4)

	byKey := make(map[string]models.AmaliyahMetricScore, len(metrics))
	for _, metric := range metrics {
		byKey[metric.Key] = metric
	}

	require.Equal(t, 3, byKey[models.AmaliyahTahajud].Available)
	require.Equal(t, 2, byKey[models.AmaliyahTahajud].Completed)
	require.InDelta(t, 66.6667, byKey[models.AmaliyahTahajud].Score, 0.001)

	require.Equal(t, 3, byKey[models.AmaliyahTilawahTadabur].Available)
	require.Equal(t, 2, byKey[models.AmaliyahTilawahTadabur].Completed)
	require.InDelta(t, 66.6667, byKey[models.AmaliyahTilawahTadabur].Score, 0.001)

	require.Equal(t, 4, byKey[models.AmaliyahDzikirPagi].Available)
	require.Equal(t, 2, byKey[models.AmaliyahDzikirPagi].Completed)
	require.Equal(t, float64(50), byKey[models.AmaliyahDzikirPagi].Score)

	require.Equal(t, 4, byKey[models.AmaliyahDzikirPetang].Available)
	require.Zero(t, byKey[models.AmaliyahDzikirPetang].Completed)
	require.Zero(t, byKey[models.AmaliyahDzikirPetang].Score)
}

func buildAmaliyahTestWorkbook(t *testing.T) []byte {
	t.Helper()

	var out bytes.Buffer
	archive := zip.NewWriter(&out)
	write := func(name, content string) {
		writer, err := archive.Create(name)
		require.NoError(t, err)
		_, err = writer.Write([]byte(content))
		require.NoError(t, err)
	}

	write("xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets><sheet name="New Amaliyah Harian" sheetId="1" r:id="rId1"/></sheets>
</workbook>`)
	write("xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/>
</Relationships>`)
	write("xl/sharedStrings.xml", `<?xml version="1.0" encoding="UTF-8"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="6" uniqueCount="6">
  <si><t>BULAN SEPTEMBER</t></si>
  <si><t>Sedekah subuh</t></si>
  <si><t>Tilawah + Tadabur</t></si>
  <si><t>Dzikir pagi</t></si>
  <si><t>Dzikir petang</t></si>
  <si><t>Tahajud</t></si>
</sst>`)

	boolCell := func(ref string, value bool) string {
		v := 0
		if value {
			v = 1
		}
		return fmt.Sprintf(`<c r="%s" t="b"><v>%d</v></c>`, ref, v)
	}

	sheet := `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>
  <row r="14"><c r="D14" t="s"><v>0</v></c></row>
  <row r="15"><c r="D15"><v>1</v></c><c r="E15"><v>2</v></c><c r="F15"><v>3</v></c><c r="G15"><v>4</v></c></row>
  <row r="16"><c r="B16" t="s"><v>1</v></c>` + boolCell("D16", true) + boolCell("E16", true) + boolCell("F16", true) + boolCell("G16", true) + `</row>
  <row r="17"><c r="B17" t="s"><v>2</v></c>` + boolCell("D17", true) + boolCell("E17", false) + boolCell("G17", true) + `</row>
  <row r="18"><c r="B18" t="s"><v>3</v></c>` + boolCell("D18", true) + boolCell("E18", true) + boolCell("F18", false) + boolCell("G18", false) + `</row>
  <row r="19"><c r="B19" t="s"><v>4</v></c>` + boolCell("D19", false) + boolCell("E19", false) + boolCell("F19", false) + boolCell("G19", false) + `</row>
  <row r="20"><c r="B20" t="s"><v>5</v></c>` + boolCell("D20", true) + boolCell("F20", true) + boolCell("G20", false) + `</row>
</sheetData></worksheet>`
	write("xl/worksheets/sheet1.xml", sheet)

	require.NoError(t, archive.Close())
	return out.Bytes()
}

func buildTrackerAmaliyahTestWorkbook(t *testing.T) []byte {
	t.Helper()

	var out bytes.Buffer
	archive := zip.NewWriter(&out)
	write := func(name, content string) {
		writer, err := archive.Create(name)
		require.NoError(t, err)
		_, err = writer.Write([]byte(content))
		require.NoError(t, err)
	}

	write("xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets><sheet name="Tracker Amaliyah Harian" sheetId="1" r:id="rId1"/></sheets>
</workbook>`)
	write("xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/>
</Relationships>`)
	write("xl/sharedStrings.xml", `<?xml version="1.0" encoding="UTF-8"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="6" uniqueCount="6">
  <si><t>BULAN SEPTEMBER</t></si>
  <si><t>Tilawah</t></si>
  <si><t>Dzikir pagi</t></si>
  <si><t>Dzikir petang</t></si>
  <si><t>Sedekah subuh</t></si>
  <si><t>Tahajjud</t></si>
</sst>`)

	boolCell := func(ref string, value bool) string {
		v := 0
		if value {
			v = 1
		}
		return fmt.Sprintf(`<c r="%s" t="b"><v>%d</v></c>`, ref, v)
	}

	sheet := `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>
  <row r="10"><c r="BR10" t="s"><v>0</v></c></row>
  <row r="11"><c r="BR11"><v>1</v></c><c r="BS11"><v>2</v></c><c r="BT11"><v>3</v></c><c r="BU11"><v>4</v></c></row>
  <row r="12"><c r="B12" t="s"><v>1</v></c>` + boolCell("BR12", true) + boolCell("BS12", false) + boolCell("BU12", true) + `</row>
  <row r="13"><c r="B13" t="s"><v>2</v></c>` + boolCell("BR13", true) + boolCell("BS13", true) + boolCell("BT13", false) + boolCell("BU13", false) + `</row>
  <row r="14"><c r="B14" t="s"><v>3</v></c>` + boolCell("BR14", false) + boolCell("BS14", false) + boolCell("BT14", false) + boolCell("BU14", false) + `</row>
  <row r="15"><c r="B15" t="s"><v>4</v></c>` + boolCell("BR15", true) + boolCell("BS15", true) + boolCell("BT15", true) + boolCell("BU15", true) + `</row>
  <row r="16"><c r="B16" t="s"><v>5</v></c>` + boolCell("BR16", true) + boolCell("BT16", true) + boolCell("BU16", false) + `</row>
</sheetData></worksheet>`
	write("xl/worksheets/sheet1.xml", sheet)

	require.NoError(t, archive.Close())
	return out.Bytes()
}
