package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMatchesAmaliyahMonthHeaderVariants(t *testing.T) {
	period := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	for _, value := range []string{
		"BULAN SEPTEMBER",
		"September 2026",
		"Bulan Sep 2026",
		"SEPT 2026",
		"09/2026",
		"2026-09",
	} {
		t.Run(value, func(t *testing.T) {
			require.True(t, matchesAmaliyahMonthHeader(value, period))
		})
	}
	require.False(t, matchesAmaliyahMonthHeader("Oktober 2026", period))
}

func TestLocateAmaliyahMonthBlockSkipsSummaryMonthLabel(t *testing.T) {
	cells := map[string]amaliyahWorksheetCell{
		"B5":  {Ref: "B5", Value: "Score September"},
		"CY10": {Ref: "CY10", Value: "September 2026"},
	}
	for day := 1; day <= 10; day++ {
		col := columnLetters(columnIndexFromCellRef("CY10") + day - 1)
		cells[fmt.Sprintf("%s11", col)] = amaliyahWorksheetCell{Ref: fmt.Sprintf("%s11", col), Value: fmt.Sprintf("%d", day)}
	}
	decode := func(cell amaliyahWorksheetCell) string { return cell.Value }
	period := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)

	columns, row := locateAmaliyahMonthBlock(cells, decode, period, 30)
	require.Equal(t, 11, row)
	require.Len(t, columns, 10)
	require.Equal(t, columnIndexFromCellRef("CY10"), columns[0])
}

func TestLocateAmaliyahMonthBlockUsesUniqueDateSequenceFallback(t *testing.T) {
	cells := map[string]amaliyahWorksheetCell{}
	start := columnIndexFromCellRef("D1")
	for day := 1; day <= 12; day++ {
		col := columnLetters(start + day - 1)
		ref := fmt.Sprintf("%s8", col)
		cells[ref] = amaliyahWorksheetCell{Ref: ref, Value: fmt.Sprintf("%d", day)}
	}
	decode := func(cell amaliyahWorksheetCell) string { return cell.Value }
	period := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)

	columns, row := locateAmaliyahMonthBlock(cells, decode, period, 30)
	require.Equal(t, 8, row)
	require.Len(t, columns, 12)
	require.Equal(t, start, columns[0])
}
