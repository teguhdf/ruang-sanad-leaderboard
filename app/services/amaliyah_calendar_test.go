package services

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/stretchr/testify/require"
)

func TestParseAmaliyahWorkbookIgnoresDayBeyondCalendarMonth(t *testing.T) {
	data := buildSeptember31ColumnWorkbook(t)

	metrics, err := parseAmaliyahWorkbook(data, "2026-09")
	require.NoError(t, err)

	byKey := map[string]models.AmaliyahMetricScore{}
	for _, metric := range metrics {
		byKey[metric.Key] = metric
	}

	for _, key := range []string{
		models.AmaliyahTahajud,
		models.AmaliyahTilawahTadabur,
		models.AmaliyahDzikirPagi,
		models.AmaliyahDzikirPetang,
	} {
		require.Equal(t, 30, byKey[key].Available)
		require.Equal(t, 30, byKey[key].Completed)
		require.Equal(t, float64(100), byKey[key].Score)
	}
}

func buildSeptember31ColumnWorkbook(t *testing.T) []byte {
	t.Helper()

	var out bytes.Buffer
	archive := zip.NewWriter(&out)
	write := func(name, content string) {
		w, err := archive.Create(name)
		require.NoError(t, err)
		_, err = w.Write([]byte(content))
		require.NoError(t, err)
	}

	write("xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<sheets><sheet name="Tracker Amaliyah Harian" sheetId="1" r:id="rId1"/></sheets></workbook>`)
	write("xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/>
</Relationships>`)
	write("xl/sharedStrings.xml", `<?xml version="1.0" encoding="UTF-8"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="5" uniqueCount="5">
<si><t>BULAN SEPTEMBER</t></si><si><t>Tilawah</t></si><si><t>Dzikir pagi</t></si><si><t>Dzikir petang</t></si><si><t>Tahajjud</t></si>
</sst>`)

	var dates strings.Builder
	var tilawah strings.Builder
	var pagi strings.Builder
	var petang strings.Builder
	var tahajud strings.Builder
	for day := 1; day <= 31; day++ {
		col := columnLetters(columnIndexFromCellRef("BR1") + day - 1)
		dates.WriteString(fmt.Sprintf(`<c r="%s11"><v>%d</v></c>`, col, day))
		for _, b := range []*strings.Builder{&tilawah, &pagi, &petang, &tahajud} {
			b.WriteString(fmt.Sprintf(`<c r="%s%d" t="b"><v>1</v></c>`, col, map[*strings.Builder]int{&tilawah:12, &pagi:13, &petang:14, &tahajud:16}[b]))
		}
	}

	sheet := `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>
<row r="10"><c r="BR10" t="s"><v>0</v></c></row>
<row r="11">` + dates.String() + `</row>
<row r="12"><c r="B12" t="s"><v>1</v></c>` + tilawah.String() + `</row>
<row r="13"><c r="B13" t="s"><v>2</v></c>` + pagi.String() + `</row>
<row r="14"><c r="B14" t="s"><v>3</v></c>` + petang.String() + `</row>
<row r="16"><c r="B16" t="s"><v>4</v></c>` + tahajud.String() + `</row>
</sheetData></worksheet>`
	write("xl/worksheets/sheet1.xml", sheet)

	require.NoError(t, archive.Close())
	return out.Bytes()
}
