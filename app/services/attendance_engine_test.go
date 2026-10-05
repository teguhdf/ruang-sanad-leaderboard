package services

import (
	"testing"
	"time"

	"github.com/maulanashalihin/laju-go/app/models"
)

func TestBuildAttendanceWeekPlanAppliesApprovedLeaveToWeeklyTarget(t *testing.T) {
	loc := time.FixedZone("WIB", 7*60*60)
	start := time.Date(2026, 10, 5, 0, 0, 0, 0, loc) // Monday
	end := start.AddDate(0, 0, 6)
	policy := models.AttendancePolicy{WeekdayTargetMinutes: 480, SaturdayTargetMinutes: 240}
	workDays := []string{"mon", "tue", "wed", "thu", "fri", "sat"}
	leave := []models.ApprovedLeavePeriod{{StartDate: "2026-10-07", EndDate: "2026-10-07"}}

	plan := buildAttendanceWeekPlan(start, end, policy, workDays, nil, leave)

	if plan.RequiredMinutes != 2160 {
		t.Fatalf("expected weekly requirement 2160 after one weekday leave, got %d", plan.RequiredMinutes)
	}
	if got := plan.Days[2]; got.DayType != "leave" || got.RequiredMinutes != 0 || got.Source != "approved_leave" {
		t.Fatalf("expected Wednesday approved leave with zero target, got %+v", got)
	}
}

func TestBuildAttendanceWeekPlanUsesRosterForSundayAndReplacementOff(t *testing.T) {
	loc := time.FixedZone("WIB", 7*60*60)
	start := time.Date(2026, 10, 5, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 0, 6)
	policy := models.AttendancePolicy{WeekdayTargetMinutes: 480, SaturdayTargetMinutes: 240}
	workDays := []string{"mon", "tue", "wed", "thu", "fri", "sat"}
	rosters := []models.AttendanceRosterDay{
		{WorkDate: "2026-10-07", DayType: "replacement_off", WorkMode: "onsite", RequiredMinutes: 0},
		{WorkDate: "2026-10-11", DayType: "work", WorkMode: "remote", RequiredMinutes: 480},
	}

	plan := buildAttendanceWeekPlan(start, end, policy, workDays, rosters, nil)

	if plan.RequiredMinutes != 2640 {
		t.Fatalf("expected replacement Sunday to preserve 2640 weekly target, got %d", plan.RequiredMinutes)
	}
	if got := plan.Days[2]; got.DayType != "replacement_off" || got.RequiredMinutes != 0 {
		t.Fatalf("expected Wednesday replacement off, got %+v", got)
	}
	if got := plan.Days[6]; got.DayType != "work" || got.RequiredMinutes != 480 || got.WorkMode != "remote" {
		t.Fatalf("expected Sunday roster work 480 remote, got %+v", got)
	}
}

func TestBuildAttendanceWeekPlanDoesNotChargeLeaveOnOffDay(t *testing.T) {
	loc := time.FixedZone("WIB", 7*60*60)
	start := time.Date(2026, 10, 5, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 0, 6)
	policy := models.AttendancePolicy{WeekdayTargetMinutes: 480, SaturdayTargetMinutes: 240}
	workDays := []string{"mon", "tue", "wed", "thu", "fri", "sat"}
	leave := []models.ApprovedLeavePeriod{{StartDate: "2026-10-11", EndDate: "2026-10-11"}}

	plan := buildAttendanceWeekPlan(start, end, policy, workDays, nil, leave)

	if plan.RequiredMinutes != 2640 {
		t.Fatalf("expected Sunday leave range not to reduce target, got %d", plan.RequiredMinutes)
	}
	if got := plan.Days[6]; got.DayType != "off" || got.RequiredMinutes != 0 {
		t.Fatalf("expected Sunday to stay off, got %+v", got)
	}
}


func TestBuildAttendanceWeekPlanIgnoresRosterRowsEqualToDefault(t *testing.T) {
	loc := time.FixedZone("WIB", 7*60*60)
	start := time.Date(2026, 10, 5, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 0, 6)
	policy := models.AttendancePolicy{WeekdayTargetMinutes: 480, SaturdayTargetMinutes: 240}
	workDays := []string{"mon", "tue", "wed", "thu", "fri", "sat"}
	rosters := []models.AttendanceRosterDay{
		{WorkDate: "2026-10-05", DayType: "work", WorkMode: "onsite", RequiredMinutes: 480},
		{WorkDate: "2026-10-10", DayType: "work", WorkMode: "onsite", RequiredMinutes: 240},
		{WorkDate: "2026-10-11", DayType: "off", WorkMode: "onsite", RequiredMinutes: 0},
	}

	plan := buildAttendanceWeekPlan(start, end, policy, workDays, rosters, nil)

	if plan.RequiredMinutes != 2640 {
		t.Fatalf("expected default-equivalent roster rows to keep 2640 target, got %d", plan.RequiredMinutes)
	}
	if got := plan.Days[0]; got.Source != "default" || got.DayType != "work" || got.RequiredMinutes != 480 {
		t.Fatalf("expected Monday to remain default, got %+v", got)
	}
	if got := plan.Days[5]; got.Source != "default" || got.DayType != "work" || got.RequiredMinutes != 240 {
		t.Fatalf("expected Saturday to remain default, got %+v", got)
	}
	if got := plan.Days[6]; got.Source != "default" || got.DayType != "off" || got.RequiredMinutes != 0 {
		t.Fatalf("expected Sunday to remain default off, got %+v", got)
	}
}

func TestBuildAttendanceWeekPlanKeepsMaterialRosterOverride(t *testing.T) {
	loc := time.FixedZone("WIB", 7*60*60)
	start := time.Date(2026, 10, 5, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 0, 6)
	policy := models.AttendancePolicy{WeekdayTargetMinutes: 480, SaturdayTargetMinutes: 240}
	workDays := []string{"mon", "tue", "wed", "thu", "fri", "sat"}
	rosters := []models.AttendanceRosterDay{
		{WorkDate: "2026-10-06", DayType: "replacement_off", WorkMode: "onsite", RequiredMinutes: 0},
		{WorkDate: "2026-10-11", DayType: "work", WorkMode: "onsite", RequiredMinutes: 480},
	}

	plan := buildAttendanceWeekPlan(start, end, policy, workDays, rosters, nil)

	if got := plan.Days[1]; got.Source != "roster" || got.DayType != "replacement_off" {
		t.Fatalf("expected Tuesday replacement-off override, got %+v", got)
	}
	if got := plan.Days[6]; got.Source != "roster" || got.DayType != "work" || got.RequiredMinutes != 480 {
		t.Fatalf("expected Sunday work override, got %+v", got)
	}
}
