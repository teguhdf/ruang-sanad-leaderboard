package models

type AttendanceRosterDay struct {
	WorkDate        string `json:"work_date"`
	DayType         string `json:"day_type"`
	ShiftTemplateID *int64 `json:"shift_template_id,omitempty"`
	WorkMode        string `json:"work_mode"`
	RequiredMinutes int    `json:"required_minutes"`
}

type ApprovedLeavePeriod struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type AttendanceDayPlan struct {
	WorkDate        string `json:"work_date"`
	DayType         string `json:"day_type"`
	ShiftTemplateID *int64 `json:"shift_template_id,omitempty"`
	WorkMode        string `json:"work_mode"`
	RequiredMinutes int    `json:"required_minutes"`
	Source          string `json:"source"`
}

type AttendanceWeekPlan struct {
	Days            []AttendanceDayPlan `json:"days"`
	RequiredMinutes int                 `json:"required_minutes"`
}

type MemberAttendanceWeek struct {
	MemberID       int64                 `json:"member_id"`
	MemberName     string                `json:"member_name"`
	Division       string                `json:"division"`
	WeekStart      string                `json:"week_start"`
	WeekEnd        string                `json:"week_end"`
	Plan           AttendanceWeekPlan    `json:"plan"`
}

type RosterDayInput struct {
	DayType         string `json:"day_type" form:"day_type"`
	ShiftTemplateID *int64 `json:"shift_template_id" form:"shift_template_id"`
	WorkMode        string `json:"work_mode" form:"work_mode"`
	RequiredMinutes int    `json:"required_minutes" form:"required_minutes"`
}
