package models

import "time"

const (
	AmaliyahTilawahTadabur = "tilawah_tadabur"
	AmaliyahDzikirPagi     = "dzikir_pagi"
	AmaliyahDzikirPetang   = "dzikir_petang"
	AmaliyahTahajud        = "tahajud"
)

type AmaliyahMetricScore struct {
	Key       string  `json:"key"`
	Label     string  `json:"label"`
	Completed int     `json:"completed"`
	Available int     `json:"available"`
	Score     float64 `json:"score"`
}

type AmaliyahMemberSnapshot struct {
	ID             int64                 `json:"id"`
	MemberID       int64                 `json:"member_id"`
	MemberName     string                `json:"member_name"`
	Division       string                `json:"division"`
	PeriodMonth    string                `json:"period_month"`
	SourceFileName string                `json:"source_file_name"`
	OverallScore   float64               `json:"overall_score"`
	Metrics        []AmaliyahMetricScore `json:"metrics"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

type AmaliyahTeamLeaderboard struct {
	Rank         int                   `json:"rank"`
	Division     string                `json:"division"`
	MemberCount  int                   `json:"member_count"`
	OverallScore float64               `json:"overall_score"`
	Metrics      []AmaliyahMetricScore `json:"metrics"`
}

type AmaliyahLeaderboard struct {
	PeriodMonth string                    `json:"period_month"`
	Teams       []AmaliyahTeamLeaderboard `json:"teams"`
	Members     []AmaliyahMemberSnapshot  `json:"members"`
	Periods     []string                  `json:"periods"`
}
