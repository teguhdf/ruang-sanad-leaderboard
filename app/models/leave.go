package models

import "time"

type LeaveEntitlement struct {
	ID             int64     `json:"id"`
	MemberID       int64     `json:"member_id"`
	MemberName     string    `json:"member_name"`
	Division       string    `json:"division"`
	Year           int       `json:"year"`
	EntitledDays   int       `json:"entitled_days"`
	CarryoverDays  int       `json:"carryover_days"`
	AdjustmentDays int       `json:"adjustment_days"`
	Note           string    `json:"note"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type LeaveBalance struct {
	MemberID       int64  `json:"member_id"`
	MemberName     string `json:"member_name"`
	Division       string `json:"division"`
	Year           int    `json:"year"`
	EntitledDays   int    `json:"entitled_days"`
	CarryoverDays  int    `json:"carryover_days"`
	AdjustmentDays int    `json:"adjustment_days"`
	ApprovedDays   int    `json:"approved_days"`
	PendingDays    int    `json:"pending_days"`
	AvailableDays  int    `json:"available_days"`
}

type LeaveRequest struct {
	ID                int64      `json:"id"`
	MemberID          int64      `json:"member_id"`
	MemberName        string     `json:"member_name"`
	Division          string     `json:"division"`
	EntitlementYear   int        `json:"entitlement_year"`
	StartDate         string     `json:"start_date"`
	EndDate           string     `json:"end_date"`
	ChargeableDays    int        `json:"chargeable_days"`
	Reason            string     `json:"reason"`
	Status            string     `json:"status"`
	ReviewNote        string     `json:"review_note"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type LeaveRequestInput struct {
	StartDate string `json:"start_date" form:"start_date"`
	EndDate   string `json:"end_date" form:"end_date"`
	Reason    string `json:"reason" form:"reason"`
}

type LeaveEntitlementInput struct {
	Year          int    `json:"year" form:"year"`
	EntitledDays int    `json:"entitled_days" form:"entitled_days"`
	Note          string `json:"note" form:"note"`
}
