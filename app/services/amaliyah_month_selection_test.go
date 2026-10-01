package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLocateAmaliyahMonthBlockChoosesSequenceNearestMatchingHeader(t *testing.T) {
	cells := map[string]amaliyahWorksheetCell{}
	put := func(ref, value string) {
		cells[ref] = amaliyahWorksheetCell{Ref: ref, Value: value}
	}

	// Several month blocks share the same date row. The selected September
	// header is near BR, while another equally long 1..30 sequence exists at D.
	// The parser must bind the header to the nearest sequence instead of using
	// whichever map entry happens to be visited first.
	put("BQ10", "BULAN SEPTEMBER")

	for day := 1; day <= 30; day++ {
		put(fmt.Sprintf("%s11", columnLetters(columnIndexFromCellRef("D1")+day-1)), fmt.Sprintf("%d", day))
		put(fmt.Sprintf("%s11", columnLetters(columnIndexFromCellRef("BR1")+day-1)), fmt.Sprintf("%d", day))
		put(fmt.Sprintf("%s11", columnLetters(columnIndexFromCellRef("DA1")+day-1)), fmt.Sprintf("%d", day))
	}

	decode := func(cell amaliyahWorksheetCell) string { return cell.Value }
	period := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 100; i++ {
		columns, row := locateAmaliyahMonthBlock(cells, decode, period, 30)
		require.Equal(t, 11, row)
		require.Len(t, columns, 30)
		require.Equal(t, columnIndexFromCellRef("BR1"), columns[0])
		require.Equal(t, columnIndexFromCellRef("CU1"), columns[29])
	}
}
