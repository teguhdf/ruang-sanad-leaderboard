package models

import "time"

type EmployeeAttendanceIdentity struct {
	OwnerUserID         int64  `json:"-"`
	MemberID            int64  `json:"member_id"`
	MemberName          string `json:"member_name"`
	Division            string `json:"division"`
	WorkMode            string `json:"work_mode"`
	ScheduleType        string `json:"schedule_type"`
	WeeklyTargetMinutes int    `json:"weekly_target_minutes"`
}

type AttendanceSession struct {
	ID              int64      `json:"id"`
	MemberID        int64      `json:"member_id"`
	WorkDate        string     `json:"work_date"`
	LocationMode    string     `json:"location_mode"`
	CheckInAt       time.Time  `json:"check_in_at"`
	CheckOutAt      *time.Time `json:"check_out_at,omitempty"`
	Status          string     `json:"status"`
	Source          string     `json:"source"`
	DurationMinutes int        `json:"duration_minutes"`
}

type AttendanceAdjustmentRequest struct {
	ID                int64      `json:"id"`
	MemberID          int64      `json:"member_id"`
	MemberName        string     `json:"member_name"`
	RelatedSessionID  *int64     `json:"related_session_id,omitempty"`
	WorkDate          string     `json:"work_date"`
	LocationMode      string     `json:"location_mode"`
	RequestedCheckIn  time.Time  `json:"requested_check_in"`
	RequestedCheckOut time.Time  `json:"requested_check_out"`
	Reason            string     `json:"reason"`
	Status            string     `json:"status"`
	ReviewNote        string     `json:"review_note"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type StartWorkInput struct {
	LocationMode string `json:"location_mode" form:"location_mode"`
}

type ForgotCheckoutInput struct {
	EndTime string `json:"end_time" form:"end_time"`
	Reason  string `json:"reason" form:"reason"`
}

type AttendanceAdjustmentInput struct {
	WorkDate     string `json:"work_date" form:"work_date"`
	LocationMode string `json:"location_mode" form:"location_mode"`
	StartTime    string `json:"start_time" form:"start_time"`
	EndTime      string `json:"end_time" form:"end_time"`
	Reason       string `json:"reason" form:"reason"`
}

type OvertimeRequestInput struct {
	WorkDate         string `json:"work_date" form:"work_date"`
	RequestedMinutes int    `json:"requested_minutes" form:"requested_minutes"`
	Reason           string `json:"reason" form:"reason"`
}

type ReviewRequestInput struct {
	Action     string `json:"action" form:"action"`
	ReviewNote string `json:"review_note" form:"review_note"`
}

type MyAttendanceSummary struct {
	Identity             EmployeeAttendanceIdentity    `json:"identity"`
	Today                string                        `json:"today"`
	TodayPlan            AttendanceDayPlan             `json:"today_plan"`
	WeekPlan             []AttendanceDayPlan           `json:"week_plan"`
	ActiveSession        *AttendanceSession            `json:"active_session,omitempty"`
	TodaySessions        []AttendanceSession           `json:"today_sessions"`
	TodayMinutes         int                           `json:"today_minutes"`
	WeekStart            string                        `json:"week_start"`
	WeekEnd              string                        `json:"week_end"`
	WeeklyMinutes        int                           `json:"weekly_minutes"`
	WeeklyTargetMinutes  int                           `json:"weekly_target_minutes"`
	WeeklyBalanceMinutes int                           `json:"weekly_balance_minutes"`
	Adjustments          []AttendanceAdjustmentRequest `json:"adjustments"`
	OvertimeRequests     []OvertimeRequest             `json:"overtime_requests"`
}
