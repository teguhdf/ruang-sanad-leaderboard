package services

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/maulanashalihin/laju-go/app/models"
)

const maxAmaliyahXMLBytes = 8 * 1024 * 1024

type amaliyahWorkbookXML struct {
	Sheets []amaliyahWorkbookSheet `xml:"sheets>sheet"`
}

type amaliyahWorkbookSheet struct {
	Name  string `xml:"name,attr"`
	RelID string `xml:"id,attr"`
}

type amaliyahRelationshipsXML struct {
	Items []amaliyahRelationship `xml:"Relationship"`
}

type amaliyahRelationship struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
}

type amaliyahSharedStringsXML struct {
	Items []amaliyahSharedString `xml:"si"`
}

type amaliyahSharedString struct {
	Text string                    `xml:"t"`
	Runs []amaliyahSharedStringRun `xml:"r"`
}

type amaliyahSharedStringRun struct {
	Text string `xml:"t"`
}

type amaliyahWorksheetXML struct {
	Rows []amaliyahWorksheetRow `xml:"sheetData>row"`
}

type amaliyahWorksheetRow struct {
	Cells []amaliyahWorksheetCell `xml:"c"`
}

type amaliyahWorksheetCell struct {
	Ref    string `xml:"r,attr"`
	Type   string `xml:"t,attr"`
	Value  string `xml:"v"`
	Inline struct {
		Text string `xml:"t"`
	} `xml:"is"`
}

var monthNamesID = map[string]string{
	"01": "JANUARI",
	"02": "FEBRUARI",
	"03": "MARET",
	"04": "APRIL",
	"05": "MEI",
	"06": "JUNI",
	"07": "JULI",
	"08": "AGUSTUS",
	"09": "SEPTEMBER",
	"10": "OKTOBER",
	"11": "NOVEMBER",
	"12": "DESEMBER",
}

var amaliyahMonthAliases = map[time.Month][]string{
	time.January:   {"januari", "january", "jan"},
	time.February:  {"februari", "february", "feb"},
	time.March:     {"maret", "march", "mar"},
	time.April:     {"april", "apr"},
	time.May:       {"mei", "may"},
	time.June:      {"juni", "june", "jun"},
	time.July:      {"juli", "july", "jul"},
	time.August:    {"agustus", "august", "agu", "aug"},
	time.September: {"september", "sept", "sep"},
	time.October:   {"oktober", "october", "okt", "oct"},
	time.November:  {"november", "nov"},
	time.December:  {"desember", "december", "des", "dec"},
}

func parseAmaliyahWorkbook(data []byte, periodMonth string) ([]models.AmaliyahMetricScore, error) {
	if len(periodMonth) != 7 {
		return nil, &AmaliyahValidationError{Message: "Periode bulan tidak valid."}
	}
	monthName, ok := monthNamesID[periodMonth[5:7]]
	if !ok {
		return nil, &AmaliyahValidationError{Message: "Periode bulan tidak valid."}
	}
	periodStart, err := time.Parse("2006-01", periodMonth)
	if err != nil {
		return nil, &AmaliyahValidationError{Message: "Periode bulan tidak valid."}
	}
	maxCalendarDays := time.Date(periodStart.Year(), periodStart.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()

	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, &AmaliyahValidationError{Message: "File bukan workbook .xlsx yang valid."}
	}

	workbookRaw, err := readAmaliyahZipEntry(archive, "xl/workbook.xml")
	if err != nil {
		return nil, &AmaliyahValidationError{Message: "Struktur workbook tidak lengkap."}
	}
	var workbook amaliyahWorkbookXML
	if err := xml.Unmarshal(workbookRaw, &workbook); err != nil {
		return nil, err
	}

	relID := ""
	for _, sheet := range workbook.Sheets {
		name := normalizeAmaliyahText(sheet.Name)
		if name == "new amaliyah harian" || strings.Contains(name, "amaliyah harian") {
			relID = sheet.RelID
			break
		}
	}
	if relID == "" && len(workbook.Sheets) == 1 {
		relID = workbook.Sheets[0].RelID
	}
	if relID == "" {
		return nil, &AmaliyahValidationError{Message: "Sheet amaliyah harian tidak ditemukan. Pastikan link workspace mengarah ke tab tracker amaliyah."}
	}

	relsRaw, err := readAmaliyahZipEntry(archive, "xl/_rels/workbook.xml.rels")
	if err != nil {
		return nil, &AmaliyahValidationError{Message: "Relasi sheet workbook tidak ditemukan."}
	}
	var relationships amaliyahRelationshipsXML
	if err := xml.Unmarshal(relsRaw, &relationships); err != nil {
		return nil, err
	}

	sheetTarget := ""
	for _, rel := range relationships.Items {
		if rel.ID == relID {
			sheetTarget = strings.TrimPrefix(rel.Target, "/")
			break
		}
	}
	if sheetTarget == "" {
		return nil, &AmaliyahValidationError{Message: "Sheet amaliyah tidak dapat dibuka."}
	}
	if !strings.HasPrefix(sheetTarget, "xl/") {
		sheetTarget = path.Join("xl", sheetTarget)
	}

	sharedStrings := make([]string, 0)
	if raw, err := readAmaliyahZipEntry(archive, "xl/sharedStrings.xml"); err == nil {
		var shared amaliyahSharedStringsXML
		if err := xml.Unmarshal(raw, &shared); err != nil {
			return nil, err
		}
		sharedStrings = make([]string, len(shared.Items))
		for i, item := range shared.Items {
			var value strings.Builder
			value.WriteString(item.Text)
			for _, run := range item.Runs {
				value.WriteString(run.Text)
			}
			sharedStrings[i] = value.String()
		}
	}

	sheetRaw, err := readAmaliyahZipEntry(archive, sheetTarget)
	if err != nil {
		return nil, &AmaliyahValidationError{Message: "Data sheet amaliyah tidak ditemukan."}
	}
	var sheet amaliyahWorksheetXML
	if err := xml.Unmarshal(sheetRaw, &sheet); err != nil {
		return nil, err
	}

	cells := make(map[string]amaliyahWorksheetCell)
	for _, row := range sheet.Rows {
		for _, cell := range row.Cells {
			if cell.Ref != "" {
				cells[cell.Ref] = cell
			}
		}
	}

	decode := func(cell amaliyahWorksheetCell) string {
		switch cell.Type {
		case "s":
			idx, err := strconv.Atoi(cell.Value)
			if err == nil && idx >= 0 && idx < len(sharedStrings) {
				return sharedStrings[idx]
			}
		case "inlineStr", "str":
			if cell.Inline.Text != "" {
				return cell.Inline.Text
			}
			return cell.Value
		default:
			return cell.Value
		}
		return ""
	}

	dayColumns, dayRow := locateAmaliyahMonthBlock(cells, decode, periodStart, maxCalendarDays)
	if len(dayColumns) == 0 || dayRow == 0 {
		return nil, &AmaliyahValidationError{Message: fmt.Sprintf("Blok bulan %s tidak ditemukan. Parser sudah mencoba nama bulan lengkap, singkatan, format bulan-tahun, dan pola tanggal.", monthName)}
	}

	metricRows := make(map[string]int, 4)
	for ref, cell := range cells {
		labelColumn := columnIndexFromCellRef(ref)
		if labelColumn < 1 || labelColumn > 4 {
			continue
		}
		row := rowIndexFromCellRef(ref)
		if row <= dayRow || row > dayRow+50 {
			continue
		}

		key := classifyAmaliyahMetricLabel(decode(cell))
		if key == "" {
			continue
		}

		hasCheckbox := false
		for _, dayColumn := range dayColumns {
			dayCell, ok := cells[fmt.Sprintf("%s%d", columnLetters(dayColumn), row)]
			if ok && dayCell.Type == "b" {
				hasCheckbox = true
				break
			}
		}
		if !hasCheckbox {
			continue
		}

		if current := metricRows[key]; current == 0 || row < current {
			metricRows[key] = row
		}
	}

	for _, key := range amaliyahMetricOrder {
		if metricRows[key] == 0 {
			return nil, &AmaliyahValidationError{Message: fmt.Sprintf("Baris %s tidak ditemukan.", amaliyahMetricLabels[key])}
		}
	}

	metrics := make([]models.AmaliyahMetricScore, 0, len(amaliyahMetricOrder))
	availableTotal := 0
	for _, key := range amaliyahMetricOrder {
		row := metricRows[key]
		metric := models.AmaliyahMetricScore{Key: key, Label: amaliyahMetricLabels[key]}
		for _, col := range dayColumns {
			cell, ok := cells[fmt.Sprintf("%s%d", columnLetters(col), row)]
			if !ok || cell.Type != "b" {
				continue
			}
			metric.Available++
			if strings.TrimSpace(cell.Value) == "1" {
				metric.Completed++
			}
		}
		if metric.Available > 0 {
			metric.Score = float64(metric.Completed) / float64(metric.Available) * 100
		}
		availableTotal += metric.Available
		metrics = append(metrics, metric)
	}

	if availableTotal == 0 {
		return nil, &AmaliyahValidationError{Message: "Tidak ada checkbox aktif pada empat amaliyah untuk bulan yang dipilih."}
	}
	return metrics, nil
}

func locateAmaliyahMonthBlock(cells map[string]amaliyahWorksheetCell, decode func(amaliyahWorksheetCell) string, period time.Time, maxDays int) ([]int, int) {
	bestColumns := []int(nil)
	bestRow := 0

	// Merged month headers often anchor far to the left of day 1. Once a month
	// label is found, search the nearby rows for a real 1..N date sequence across
	// the entire row instead of assuming day 1 is within a few columns of the
	// header anchor.
	for ref, cell := range cells {
		headerRow := rowIndexFromCellRef(ref)
		if headerRow <= 0 || headerRow > 200 || !matchesAmaliyahMonthHeader(decode(cell), period) {
			continue
		}
		candidateColumns, candidateRow := findBestAmaliyahDaySequenceInRows(cells, decode, headerRow, headerRow+6, maxDays)
		if len(candidateColumns) > len(bestColumns) {
			bestColumns = candidateColumns
			bestRow = candidateRow
		}
	}
	if len(bestColumns) > 0 {
		return bestColumns, bestRow
	}

	// Structural fallback for renamed/custom templates: if the exported sheet
	// contains exactly one plausible 1..N date sequence, use it even when the
	// month label itself was removed or formatted unexpectedly.
	type block struct {
		columns []int
		row     int
		col     int
	}
	blocks := make([]block, 0)
	for ref, cell := range cells {
		row := rowIndexFromCellRef(ref)
		if row <= 0 || row > 200 {
			continue
		}
		value, err := strconv.ParseFloat(strings.TrimSpace(decode(cell)), 64)
		if err != nil || int(value) != 1 {
			continue
		}
		col := columnIndexFromCellRef(ref)
		candidate := readAmaliyahDaySequence(cells, decode, col, row, maxDays)
		if len(candidate) < 7 {
			continue
		}
		blocks = append(blocks, block{columns: candidate, row: row, col: col})
	}
	if len(blocks) == 1 {
		return blocks[0].columns, blocks[0].row
	}

	// If several sequences exist, prefer a unique one whose length exactly
	// matches the selected calendar month. Otherwise fail rather than guessing.
	exact := make([]block, 0, len(blocks))
	for _, item := range blocks {
		if len(item.columns) == maxDays {
			exact = append(exact, item)
		}
	}
	if len(exact) == 1 {
		return exact[0].columns, exact[0].row
	}
	return nil, 0
}

func findBestAmaliyahDaySequenceInRows(cells map[string]amaliyahWorksheetCell, decode func(amaliyahWorksheetCell) string, startRow, endRow, maxDays int) ([]int, int) {
	bestColumns := []int(nil)
	bestRow := 0
	for ref, cell := range cells {
		row := rowIndexFromCellRef(ref)
		if row < startRow || row > endRow {
			continue
		}
		value, err := strconv.ParseFloat(strings.TrimSpace(decode(cell)), 64)
		if err != nil || int(value) != 1 {
			continue
		}
		startColumn := columnIndexFromCellRef(ref)
		candidate := readAmaliyahDaySequence(cells, decode, startColumn, row, maxDays)
		if len(candidate) >= 7 && len(candidate) > len(bestColumns) {
			bestColumns = candidate
			bestRow = row
		}
	}
	return bestColumns, bestRow
}

func readAmaliyahDaySequence(cells map[string]amaliyahWorksheetCell, decode func(amaliyahWorksheetCell) string, startColumn, row, maxDays int) []int {
	candidate := make([]int, 0, maxDays)
	for expectedDay := 1; expectedDay <= maxDays; expectedDay++ {
		col := startColumn + expectedDay - 1
		cell, ok := cells[fmt.Sprintf("%s%d", columnLetters(col), row)]
		if !ok {
			break
		}
		value, err := strconv.ParseFloat(strings.TrimSpace(decode(cell)), 64)
		if err != nil || int(value) != expectedDay {
			break
		}
		candidate = append(candidate, col)
	}
	return candidate
}

func matchesAmaliyahMonthHeader(value string, period time.Time) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.NewReplacer(
		"/", " ", "-", " ", "_", " ", ".", " ", ",", " ", ":", " ", "(", " ", ")", " ",
	).Replace(normalized)
	fields := strings.Fields(normalized)
	if len(fields) == 0 {
		return false
	}

	containsToken := func(token string) bool {
		for _, field := range fields {
			if field == token {
				return true
			}
		}
		return false
	}
	for _, alias := range amaliyahMonthAliases[period.Month()] {
		if containsToken(alias) {
			return true
		}
	}

	monthNumber := fmt.Sprintf("%02d", int(period.Month()))
	year := strconv.Itoa(period.Year())
	if containsToken(monthNumber) && containsToken(year) {
		return true
	}
	compact := strings.ReplaceAll(normalized, " ", "")
	return strings.Contains(compact, year+monthNumber) || strings.Contains(compact, monthNumber+year)
}

func classifyAmaliyahMetricLabel(value string) string {
	normalized := normalizeAmaliyahText(value)
	normalized = strings.NewReplacer(
		"/", " ",
		"-", " ",
		"_", " ",
		"+", " ",
		"&", " ",
		"(", " ",
		")", " ",
		":", " ",
	).Replace(normalized)
	normalized = normalizeAmaliyahText(normalized)

	hasDzikir := strings.Contains(normalized, "dzikir") || strings.Contains(normalized, "zikir") || strings.Contains(normalized, "dzizkir")
	hasPetang := strings.Contains(normalized, "petang") || strings.Contains(normalized, "pertang") || strings.Contains(normalized, "sore") || strings.Contains(normalized, "malam")
	switch {
	case strings.Contains(normalized, "tilawah"):
		return models.AmaliyahTilawahTadabur
	case hasDzikir && strings.Contains(normalized, "pagi"):
		return models.AmaliyahDzikirPagi
	case hasDzikir && hasPetang:
		return models.AmaliyahDzikirPetang
	case strings.Contains(normalized, "tahajud") || strings.Contains(normalized, "tahajjud"):
		return models.AmaliyahTahajud
	default:
		return ""
	}
}

func readAmaliyahZipEntry(archive *zip.Reader, name string) ([]byte, error) {
	for _, file := range archive.File {
		if file.Name != name {
			continue
		}
		if file.UncompressedSize64 > maxAmaliyahXMLBytes {
			return nil, fmt.Errorf("xlsx entry too large: %s", name)
		}
		reader, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer reader.Close()

		data, err := io.ReadAll(io.LimitReader(reader, maxAmaliyahXMLBytes+1))
		if err != nil {
			return nil, err
		}
		if len(data) > maxAmaliyahXMLBytes {
			return nil, fmt.Errorf("xlsx entry too large: %s", name)
		}
		return data, nil
	}
	return nil, fmt.Errorf("xlsx entry not found: %s", name)
}

func normalizeAmaliyahText(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}

func columnIndexFromCellRef(ref string) int {
	column := 0
	for _, char := range ref {
		switch {
		case char >= 'A' && char <= 'Z':
			column = column*26 + int(char-'A'+1)
		case char >= 'a' && char <= 'z':
			column = column*26 + int(char-'a'+1)
		default:
			return column
		}
	}
	return column
}

func rowIndexFromCellRef(ref string) int {
	index := 0
	for index < len(ref) {
		char := ref[index]
		if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') {
			index++
			continue
		}
		break
	}
	row, _ := strconv.Atoi(ref[index:])
	return row
}

func columnLetters(column int) string {
	result := ""
	for column > 0 {
		column--
		result = string(rune('A'+column%26)) + result
		column /= 26
	}
	return result
}
