package services

import (
	"context"
	"strings"
	"time"

	"github.com/maulanashalihin/laju-go/app/models"
)

func (s *EmployeeAttendanceService) buildAttendanceWeekPlan(ctx context.Context, identity models.EmployeeAttendanceIdentity, start, end time.Time) (models.AttendanceWeekPlan, error) {
	policy, err := s.querier.GetAttendancePolicy(ctx, identity.OwnerUserID)
	if err != nil {
		return models.AttendanceWeekPlan{}, err
	}
	workDays, err := s.querier.GetMemberWorkDaysForLeave(ctx, identity.OwnerUserID, identity.MemberID)
	if err != nil {
		return models.AttendanceWeekPlan{}, err
	}
	startDate := start.Format("2006-01-02")
	endDate := end.Format("2006-01-02")
	rosters, err := s.querier.ListRostersForMemberRange(ctx, identity.OwnerUserID, identity.MemberID, startDate, endDate)
	if err != nil {
		return models.AttendanceWeekPlan{}, err
	}
	leavePeriods, err := s.querier.ListApprovedLeavePeriodsForMemberRange(ctx, identity.OwnerUserID, identity.MemberID, startDate, endDate)
	if err != nil {
		return models.AttendanceWeekPlan{}, err
	}
	return buildAttendanceWeekPlan(start, end, policy, workDays, identity.WorkMode, rosters, leavePeriods), nil
}

func buildAttendanceWeekPlan(start, end time.Time, policy models.AttendancePolicy, workDays []string, defaultWorkMode string, rosters []models.AttendanceRosterDay, leavePeriods []models.ApprovedLeavePeriod) models.AttendanceWeekPlan {
	workDaySet := make(map[string]bool, len(workDays))
	for _, day := range workDays {
		workDaySet[strings.ToLower(strings.TrimSpace(day))] = true
	}

	rosterByDate := make(map[string]models.AttendanceRosterDay, len(rosters))
	for _, roster := range rosters {
		rosterByDate[roster.WorkDate] = roster
	}

	leaveDates := make(map[string]bool)
	for _, period := range leavePeriods {
		periodStart, err := time.ParseInLocation("2006-01-02", period.StartDate, start.Location())
		if err != nil {
			continue
		}
		periodEnd, err := time.ParseInLocation("2006-01-02", period.EndDate, start.Location())
		if err != nil {
			continue
		}
		for day := periodStart; !day.After(periodEnd); day = day.AddDate(0, 0, 1) {
			leaveDates[day.Format("2006-01-02")] = true
		}
	}

	defaultMode := normalizeAttendanceWorkMode(defaultWorkMode)
	plan := models.AttendanceWeekPlan{Days: make([]models.AttendanceDayPlan, 0, 7)}
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		date := day.Format("2006-01-02")
		item := models.AttendanceDayPlan{
			WorkDate: date,
			DayType:  "off",
			WorkMode: defaultMode,
			Source:   "default",
		}

		defaultIsWork := workDaySet[attendanceWeekdayKey(day.Weekday())]
		if defaultIsWork {
			item.DayType = "work"
			item.RequiredMinutes = defaultAttendanceMinutes(day.Weekday(), policy)
		}

		if roster, ok := rosterByDate[date]; ok && !rosterMatchesAttendanceDefault(roster, day, policy, defaultIsWork, defaultMode) {
			item.DayType = normalizeAttendanceDayType(roster.DayType)
			item.ShiftTemplateID = roster.ShiftTemplateID
			item.WorkMode = strings.ToLower(strings.TrimSpace(roster.WorkMode))
			item.Source = "roster"
			if item.DayType == "work" {
				item.RequiredMinutes = maxInt(roster.RequiredMinutes, 0)
			} else {
				item.RequiredMinutes = 0
			}
		}

		if leaveDates[date] && item.DayType == "work" && item.RequiredMinutes > 0 {
			item.DayType = "leave"
			item.RequiredMinutes = 0
			item.Source = "approved_leave"
		}

		plan.RequiredMinutes += item.RequiredMinutes
		plan.Days = append(plan.Days, item)
	}
	return plan
}

func rosterMatchesAttendanceDefault(roster models.AttendanceRosterDay, day time.Time, policy models.AttendancePolicy, defaultIsWork bool, defaultWorkMode string) bool {
	dayType := normalizeAttendanceDayType(roster.DayType)
	if !defaultIsWork {
		return dayType == "off" && roster.ShiftTemplateID == nil && roster.RequiredMinutes == 0
	}
	return dayType == "work" &&
		roster.ShiftTemplateID == nil &&
		strings.EqualFold(strings.TrimSpace(roster.WorkMode), normalizeAttendanceWorkMode(defaultWorkMode)) &&
		roster.RequiredMinutes == defaultAttendanceMinutes(day.Weekday(), policy)
}

func normalizeAttendanceWorkMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "remote":
		return "remote"
	case "hybrid":
		return "hybrid"
	default:
		return "onsite"
	}
}

func defaultAttendanceMinutes(day time.Weekday, policy models.AttendancePolicy) int {
	if day == time.Saturday {
		return maxInt(policy.SaturdayTargetMinutes, 0)
	}
	return maxInt(policy.WeekdayTargetMinutes, 0)
}

func normalizeAttendanceDayType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "work":
		return "work"
	case "replacement_off":
		return "replacement_off"
	case "holiday":
		return "holiday"
	case "sick":
		return "sick"
	case "leave":
		return "leave"
	default:
		return "off"
	}
}

func attendanceWeekdayKey(day time.Weekday) string {
	switch day {
	case time.Monday:
		return "mon"
	case time.Tuesday:
		return "tue"
	case time.Wednesday:
		return "wed"
	case time.Thursday:
		return "thu"
	case time.Friday:
		return "fri"
	case time.Saturday:
		return "sat"
	default:
		return "sun"
	}
}

func maxInt(value, floor int) int {
	if value < floor {
		return floor
	}
	return value
}
