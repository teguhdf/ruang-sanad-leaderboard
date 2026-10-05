package models

import "time"

type AttendancePolicy struct {
	ID                        int64     `json:"id"`
	OwnerUserID               int64     `json:"-"`
	WeekdayTargetMinutes      int       `json:"weekday_target_minutes"`
	SaturdayTargetMinutes     int       `json:"saturday_target_minutes"`
	WeeklyTargetMinutes       int       `json:"weekly_target_minutes"`
	LateToleranceMinutes      int       `json:"late_tolerance_minutes"`
	EarlyLeaveToleranceMinutes int      `json:"early_leave_tolerance_minutes"`
	AllowWeeklyCarryover      bool      `json:"allow_weekly_carryover"`
	OvertimeRequiresApproval  bool      `json:"overtime_requires_approval"`
	WeekStartsOn              int       `json:"week_starts_on"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

type MemberAttendanceProfile struct {
	ID                   int64  `json:"id"`
	MemberID             int64  `json:"member_id"`
	MemberName           string `json:"member_name"`
	Division             string `json:"division"`
	LinkedUserID         *int64 `json:"linked_user_id,omitempty"`
	WorkMode             string `json:"work_mode"`
	ScheduleType         string `json:"schedule_type"`
	RequiredDailyMinutes int    `json:"required_daily_minutes"`
	WeeklyTargetMinutes  int    `json:"weekly_target_minutes"`
	AttendanceEnabled    bool   `json:"attendance_enabled"`
}

type MemberAttendanceProfileInput struct {
	WorkMode     string `json:"work_mode" form:"work_mode"`
	ScheduleType string `json:"schedule_type" form:"schedule_type"`
}

type ShiftSegment struct {
	ID              int64   `json:"id"`
	SortOrder       int     `json:"sort_order"`
	LocationMode    string  `json:"location_mode"`
	StartTime       *string `json:"start_time,omitempty"`
	EndTime         *string `json:"end_time,omitempty"`
	TargetMinutes   int     `json:"target_minutes"`
	Flexible        bool    `json:"is_flexible"`
}

type ShiftTemplate struct {
	ID              int64          `json:"id"`
	Code            string         `json:"code"`
	Name            string         `json:"name"`
	RequiredMinutes int            `json:"required_minutes"`
	Description     string         `json:"description"`
	Active          bool           `json:"is_active"`
	Segments        []ShiftSegment `json:"segments"`
}

type AttendanceFoundation struct {
	Policy          AttendancePolicy          `json:"policy"`
	Profiles        []MemberAttendanceProfile `json:"profiles"`
	ShiftTemplates  []ShiftTemplate           `json:"shift_templates"`
	RosterWeekStart string                    `json:"roster_week_start"`
	RosterWeekEnd   string                    `json:"roster_week_end"`
	RosterWeeks     []MemberAttendanceWeek    `json:"roster_weeks"`
}

type OvertimeRequest struct {
	ID               int64      `json:"id"`
	MemberID         int64      `json:"member_id"`
	WorkDate         string     `json:"work_date"`
	RequestedMinutes int        `json:"requested_minutes"`
	Reason           string     `json:"reason"`
	Status           string     `json:"status"`
	ApprovedByUserID *int64     `json:"approved_by_user_id,omitempty"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty"`
	ReviewNote       string     `json:"review_note"`
	CreatedAt        time.Time  `json:"created_at"`
}
