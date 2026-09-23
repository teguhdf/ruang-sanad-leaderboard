package models

import "time"

type Member struct {
	ID             int64     `json:"id"`
	OwnerUserID    int64     `json:"-"`
	Name           string    `json:"name"`
	Division       string    `json:"division"`
	WorkDays       []string  `json:"work_days"`
	StartTime      string    `json:"start_time"`
	FingerprintID  string    `json:"fingerprint_id"`
	GoogleSheetURL string    `json:"google_sheet_url"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type MemberInput struct {
	Name           string   `json:"name" form:"name"`
	Division       string   `json:"division" form:"division"`
	WorkDays       []string `json:"work_days" form:"work_days"`
	StartTime      string   `json:"start_time" form:"start_time"`
	FingerprintID  string   `json:"fingerprint_id" form:"fingerprint_id"`
	GoogleSheetURL string   `json:"google_sheet_url" form:"google_sheet_url"`
}
