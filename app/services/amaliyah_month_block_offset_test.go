package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLocateAmaliyahMonthBlockWithMergedHeaderFarFromDayOne(t *testing.T) {
	cells := map[string]amaliyahWorksheetCell{}

	put := func(ref, value string) {
		cells[ref] = amaliyahWorksheetCell{Ref: ref, Value: value}
	}

	// Mirrors workspaces where a merged month title is anchored far to the left
	// of the actual date columns, e.g. header at AJ10 while day 1 starts at BQ11.
	put("AJ10", "BULAN SEPTEMBER")
	for day := 1; day <= 30; day++ {
		put(fmt.Sprintf("%s11", columnLetters(columnIndexFromCellRef("BQ1")+day-1)), fmt.Sprintf("%d", day))
	}

	// Add another plausible sequence elsewhere so the global structural fallback
	// is ambiguous. The month-header-local search must still select row 11.
	for day := 1; day <= 30; day++ {
		put(fmt.Sprintf("%s60", columnLetters(columnIndexFromCellRef("D1")+day-1)), fmt.Sprintf("%d", day))
	}

	decode := func(cell amaliyahWorksheetCell) string { return cell.Value }
	period := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)

	columns, row := locateAmaliyahMonthBlock(cells, decode, period, 30)
	require.Equal(t, 11, row)
	require.Len(t, columns, 30)
	require.Equal(t, columnIndexFromCellRef("BQ1"), columns[0])
	require.Equal(t, columnIndexFromCellRef("CT1"), columns[29])
}
