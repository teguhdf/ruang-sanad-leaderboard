package models

import "time"

type FingerprintCorrectionInput struct {
	FingerprintID string `json:"fingerprint_id"`
	Date          string `json:"date"`
	CheckIn       string `json:"check_in"`
	CheckOut      string `json:"check_out"`
	Reason        string `json:"reason"`
}

type FingerprintImport struct {
	ID             int64     `json:"id"`
	FileName       string    `json:"file_name"`
	FileSHA256     string    `json:"-"`
	PeriodStart    string    `json:"period_start"`
	PeriodEnd      string    `json:"period_end"`
	FilePeople     int       `json:"file_people"`
	Matched        int       `json:"matched"`
	Unmatched      int       `json:"unmatched"`
	TotalSlots     int       `json:"total_slots"`
	TwoTimes       int       `json:"two_times"`
	OneTime        int       `json:"one_time"`
	Empty          int       `json:"empty"`
	MultipleTimes  int       `json:"multiple_times"`
	CorrectedCount int       `json:"corrected_count"`
	CreatedAt      time.Time `json:"created_at"`
}

type FingerprintRecord struct {
	MemberID         *int64
	FingerprintID     string
	FileName          string
	Department        string
	AttendanceDate    string
	Day               int
	RawValue          string
	RawCheckIn        string
	RawCheckOut       string
	CorrectedCheckIn  string
	CorrectedCheckOut string
	CorrectionReason  string
	IsCorrected       bool
	ReadState         string
}

type FingerprintImportSave struct {
	OwnerUserID int64
	Import      FingerprintImport
	Records     []FingerprintRecord
}

type FingerprintImportResult struct {
	Import    FingerprintImport `json:"import"`
	Duplicate bool              `json:"duplicate"`
}
