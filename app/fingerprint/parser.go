package fingerprint

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	xls "github.com/Standard-Innovation/xls"
)

var (
	ErrInvalidWorkbook = errors.New("invalid fingerprint workbook")
	periodPattern       = regexp.MustCompile(`(?i)tanggal\s+kehadiran\s*:\s*(\d{4}-\d{2}-\d{2})\s*[~～]\s*(\d{4}-\d{2}-\d{2})`)
	timePattern         = regexp.MustCompile(`\b(?:[01]\d|2[0-3]):[0-5]\d\b`)
)

type Preview struct {
	FileName    string          `json:"file_name"`
	PeriodStart string          `json:"period_start"`
	PeriodEnd   string          `json:"period_end"`
	People      []PersonPreview `json:"people"`
	Summary     Summary         `json:"summary"`
}

type Summary struct {
	FilePeople int `json:"file_people"`
	Matched    int `json:"matched"`
	Unmatched  int `json:"unmatched"`
	TotalSlots int `json:"total_slots"`
	TwoTimes   int `json:"two_times"`
	OneTime    int `json:"one_time"`
	Empty      int `json:"empty"`
	Multiple   int `json:"multiple_times"`
}

type PersonPreview struct {
	FingerprintID string       `json:"fingerprint_id"`
	FileName      string       `json:"file_name"`
	Department    string       `json:"department"`
	Matched       bool         `json:"matched"`
	MemberID      int64        `json:"member_id,omitempty"`
	MemberName    string       `json:"member_name,omitempty"`
	Days          []DayPreview `json:"days"`
	Summary       PersonSummary `json:"summary"`
}

type PersonSummary struct {
	TwoTimes int `json:"two_times"`
	OneTime  int `json:"one_time"`
	Empty    int `json:"empty"`
	Multiple int `json:"multiple_times"`
}

type DayPreview struct {
	Day       int      `json:"day"`
	Date      string   `json:"date"`
	RawValue  string   `json:"raw_value"`
	Times     []string `json:"times"`
	ReadState string   `json:"read_state"`
}

func ParseXLS(data []byte) (*Preview, error) {
	if len(data) < 8 || !bytes.Equal(data[:8], []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}) {
		return nil, ErrInvalidWorkbook
	}

	book, err := xls.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidWorkbook, err)
	}
	if book.NumSheets() != 1 {
		return nil, fmt.Errorf("%w: expected one worksheet", ErrInvalidWorkbook)
	}

	// ReadAllCells is intentional here. Some fingerprint exports omit BIFF ROW
	// records for rows that still contain cells. In that case Row.FirstCol and
	// Row.LastCol are both zero in the xls library, so scanning through Row()
	// silently misses User ID, name, day numbers, and punch cells.
	cells := book.ReadAllCells(5000)
	if len(cells) == 0 {
		return nil, fmt.Errorf("%w: worksheet is empty", ErrInvalidWorkbook)
	}

	startDate, endDate := findPeriod(cells)
	preview := &Preview{
		PeriodStart: formatDate(startDate),
		PeriodEnd:   formatDate(endDate),
		People:      make([]PersonPreview, 0),
	}

	for rowIndex := 0; rowIndex < len(cells); rowIndex++ {
		row := cells[rowIndex]

		fingerprintID := findMetadataValue(row, "user id")
		name := findMetadataValue(row, "nama")
		if fingerprintID == "" || name == "" {
			continue
		}

		person := PersonPreview{
			FingerprintID: fingerprintID,
			FileName:      name,
			Department:    findMetadataValue(row, "departemen"),
			Days:          make([]DayPreview, 0, 31),
		}

		dayRowIndex := rowIndex + 1
		punchRowIndex := rowIndex + 2
		if dayRowIndex >= len(cells) {
			continue
		}

		dayRow := cells[dayRowIndex]
		for col, dayValue := range dayRow {
			day, ok := parseDayNumber(dayValue)
			if !ok || day < 1 || day > 31 {
				continue
			}

			raw := cellValue(cells, punchRowIndex, col)
			times := parsePunchTimes(raw)
			readState := "empty"
			switch len(times) {
			case 0:
				person.Summary.Empty++
				preview.Summary.Empty++
			case 1:
				readState = "one_time"
				person.Summary.OneTime++
				preview.Summary.OneTime++
			case 2:
				readState = "two_times"
				person.Summary.TwoTimes++
				preview.Summary.TwoTimes++
			default:
				readState = "multiple_times"
				person.Summary.Multiple++
				preview.Summary.Multiple++
			}

			person.Days = append(person.Days, DayPreview{
				Day:       day,
				Date:      dateForDay(startDate, day),
				RawValue:  strings.TrimRight(raw, " \t\r\n"),
				Times:     times,
				ReadState: readState,
			})
			preview.Summary.TotalSlots++
		}

		preview.People = append(preview.People, person)
	}

	preview.Summary.FilePeople = len(preview.People)
	if len(preview.People) == 0 {
		return nil, fmt.Errorf("%w: no fingerprint user blocks found", ErrInvalidWorkbook)
	}

	return preview, nil
}

func findPeriod(cells [][]string) (time.Time, time.Time) {
	for _, row := range cells {
		for _, value := range row {
			value = strings.TrimSpace(value)
			match := periodPattern.FindStringSubmatch(value)
			if len(match) != 3 {
				continue
			}
			start, startErr := time.Parse("2006-01-02", match[1])
			end, endErr := time.Parse("2006-01-02", match[2])
			if startErr == nil && endErr == nil {
				return start, end
			}
		}
	}
	return time.Time{}, time.Time{}
}

func findMetadataValue(row []string, marker string) string {
	marker = strings.ToLower(marker)
	for col, value := range row {
		label := normalizeLabel(value)
		if strings.Contains(label, marker) {
			if col+1 < len(row) {
				return strings.TrimSpace(row[col+1])
			}
			return ""
		}
	}
	return ""
}

func cellValue(cells [][]string, row, col int) string {
	if row < 0 || row >= len(cells) || col < 0 || col >= len(cells[row]) {
		return ""
	}
	return cells[row][col]
}

func normalizeLabel(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "：", ":")
	return value
}

func parseDayNumber(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if integer, err := strconv.Atoi(value); err == nil {
		return integer, true
	}
	number, err := strconv.ParseFloat(value, 64)
	if err != nil || number != float64(int(number)) {
		return 0, false
	}
	return int(number), true
}

func parsePunchTimes(raw string) []string {
	matches := timePattern.FindAllString(raw, -1)
	if matches == nil {
		return []string{}
	}
	return matches
}

func dateForDay(start time.Time, day int) string {
	if start.IsZero() {
		return ""
	}
	date := time.Date(start.Year(), start.Month(), day, 0, 0, 0, 0, start.Location())
	return date.Format("2006-01-02")
}

func formatDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}
