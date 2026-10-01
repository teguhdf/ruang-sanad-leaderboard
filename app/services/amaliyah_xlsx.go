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
	if relID == "" {
		return nil, &AmaliyahValidationError{Message: "Sheet amaliyah harian tidak ditemukan. Pastikan workspace memakai template tim terbaru."}
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
		case "inlineStr":
			return cell.Inline.Text
		default:
			return cell.Value
		}
		return ""
	}

	// Workspace templates have evolved over time. Do not pin the month header to
	// one row: locate the selected month anywhere in the amaliyah sheet, then
	// discover the date row directly below it.
	startColumn := 0
	headerRow := 0
	fallbackColumn := 0
	fallbackRow := 0
	targetHeader := normalizeAmaliyahText("BULAN " + monthName)
	targetMonth := normalizeAmaliyahText(monthName)
	for ref, cell := range cells {
		row := rowIndexFromCellRef(ref)
		if row <= 0 || row > 100 {
			continue
		}
		value := normalizeAmaliyahText(decode(cell))
		if value == targetHeader {
			startColumn = columnIndexFromCellRef(ref)
			headerRow = row
			break
		}
		if fallbackColumn == 0 && value == targetMonth {
			fallbackColumn = columnIndexFromCellRef(ref)
			fallbackRow = row
		}
	}
	if startColumn == 0 {
		startColumn = fallbackColumn
		headerRow = fallbackRow
	}
	if startColumn == 0 || headerRow == 0 {
		return nil, &AmaliyahValidationError{Message: fmt.Sprintf("Kolom bulan %s tidak ditemukan di sheet amaliyah.", monthName)}
	}

	// Some templates keep a 31st column even for 30-day months. The calendar is
	// the hard upper bound; within valid dates, only actual checkbox cells count
	// toward the dynamic denominator.
	dayColumns := make([]int, 0, maxCalendarDays)
	dayRow := 0
	for candidateRow := headerRow + 1; candidateRow <= headerRow+5; candidateRow++ {
		candidate := make([]int, 0, maxCalendarDays)
		expectedDay := 1
		for col := startColumn; col < startColumn+maxCalendarDays; col++ {
			cell, ok := cells[fmt.Sprintf("%s%d", columnLetters(col), candidateRow)]
			if !ok {
				break
			}
			value, err := strconv.ParseFloat(strings.TrimSpace(decode(cell)), 64)
			if err != nil || int(value) != expectedDay {
				break
			}
			candidate = append(candidate, col)
			expectedDay++
		}
		if len(candidate) > len(dayColumns) {
			dayColumns = candidate
			dayRow = candidateRow
		}
	}
	if len(dayColumns) == 0 || dayRow == 0 {
		return nil, &AmaliyahValidationError{Message: "Kolom tanggal untuk bulan yang dipilih tidak ditemukan."}
	}

	// Metric labels are allowed to move across the left-side label columns and
	// may use common wording variants such as "Dzikir Sore" or "Zikir Petang".
	// A candidate row is accepted only when it actually contains checkbox cells
	// inside the selected month, preventing unrelated text from being matched.
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

	hasDzikir := strings.Contains(normalized, "dzikir") || strings.Contains(normalized, "zikir")
	switch {
	case strings.Contains(normalized, "tilawah"):
		return models.AmaliyahTilawahTadabur
	case hasDzikir && strings.Contains(normalized, "pagi"):
		return models.AmaliyahDzikirPagi
	case hasDzikir && (strings.Contains(normalized, "petang") || strings.Contains(normalized, "sore") || strings.Contains(normalized, "malam")):
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
