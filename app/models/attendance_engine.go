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
	WorkMode        string `json:"work_mode"`
	RequiredMinutes int    `json:"required_minutes"`
	Source          string `json:"source"`
}

type AttendanceWeekPlan struct {
	Days            []AttendanceDayPlan `json:"days"`
	RequiredMinutes int                 `json:"required_minutes"`
}
