package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/maulanashalihin/laju-go/app/queries"
)

var (
	ErrAmaliyahInvalidPeriod = errors.New("periode amaliyah tidak valid")
	ErrAmaliyahInvalidFile   = errors.New("workspace amaliyah tidak valid")
)

type AmaliyahValidationError struct {
	Message string
}

func (e *AmaliyahValidationError) Error() string { return e.Message }

var amaliyahMetricOrder = []string{
	models.AmaliyahTahajud,
	models.AmaliyahTilawahTadabur,
	models.AmaliyahDzikirPagi,
	models.AmaliyahDzikirPetang,
}

var amaliyahMetricLabels = map[string]string{
	models.AmaliyahTahajud:        "Tahajud",
	models.AmaliyahTilawahTadabur: "Tilawah + Tadabur",
	models.AmaliyahDzikirPagi:     "Dzikir Pagi",
	models.AmaliyahDzikirPetang:   "Dzikir Petang",
}

func (s *MemberService) ImportAmaliyahWorkspace(ownerUserID, memberID int64, periodMonth, fileName string, data []byte) error {
	periodMonth = strings.TrimSpace(periodMonth)
	if !validPeriodMonth(periodMonth) {
		return ErrAmaliyahInvalidPeriod
	}
	if len(data) == 0 {
		return ErrAmaliyahInvalidFile
	}

	metrics, err := parseAmaliyahWorkbook(data, periodMonth)
	if err != nil {
		var validationErr *AmaliyahValidationError
		if errors.As(err, &validationErr) {
			return err
		}
		return &AmaliyahValidationError{Message: "Workspace tidak dapat dibaca. Pastikan file .xlsx memakai template tim terbaru."}
	}

	overallScore := averageAvailableMetrics(metrics)
	hash := sha256.Sum256(data)

	if err := s.querier.SaveAmaliyahSnapshot(
		context.Background(),
		ownerUserID,
		memberID,
		periodMonth,
		fileName,
		hex.EncodeToString(hash[:]),
		overallScore,
		metrics,
	); err != nil {
		if errors.Is(err, queries.ErrAmaliyahMemberNotFound) {
			return &AmaliyahValidationError{Message: "Anggota tidak ditemukan atau bukan milik akun ini."}
		}
		return err
	}
	return nil
}

func (s *MemberService) GetAmaliyahLeaderboard(ownerUserID int64, periodMonth string) (*models.AmaliyahLeaderboard, error) {
	if !validPeriodMonth(periodMonth) {
		return nil, ErrAmaliyahInvalidPeriod
	}

	members, err := s.querier.ListAmaliyahMemberSnapshots(context.Background(), ownerUserID, periodMonth)
	if err != nil {
		return nil, err
	}
	periods, err := s.querier.ListAmaliyahPeriods(context.Background(), ownerUserID)
	if err != nil {
		return nil, err
	}

	for i := range members {
		members[i].Metrics = normalizeMetricScores(members[i].Metrics)
	}

	return &models.AmaliyahLeaderboard{
		PeriodMonth: periodMonth,
		Teams:       buildTeamLeaderboard(members),
		Members:     members,
		Periods:     periods,
	}, nil
}

func validPeriodMonth(value string) bool {
	if len(value) != 7 {
		return false
	}
	_, err := time.Parse("2006-01", value)
	return err == nil
}

func averageAvailableMetrics(metrics []models.AmaliyahMetricScore) float64 {
	var total float64
	count := 0
	for _, metric := range metrics {
		if metric.Available <= 0 {
			continue
		}
		total += metric.Score
		count++
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func normalizeMetricScores(metrics []models.AmaliyahMetricScore) []models.AmaliyahMetricScore {
	byKey := make(map[string]models.AmaliyahMetricScore, len(metrics))
	for _, metric := range metrics {
		metric.Label = amaliyahMetricLabels[metric.Key]
		byKey[metric.Key] = metric
	}

	ordered := make([]models.AmaliyahMetricScore, 0, len(amaliyahMetricOrder))
	for _, key := range amaliyahMetricOrder {
		metric, ok := byKey[key]
		if !ok {
			metric = models.AmaliyahMetricScore{Key: key, Label: amaliyahMetricLabels[key]}
		}
		metric.Label = amaliyahMetricLabels[key]
		ordered = append(ordered, metric)
	}
	return ordered
}

func buildTeamLeaderboard(members []models.AmaliyahMemberSnapshot) []models.AmaliyahTeamLeaderboard {
	type metricAccumulator struct {
		scoreSum     float64
		memberCount  int
		completedSum int
		availableSum int
	}
	type teamAccumulator struct {
		division    string
		memberCount int
		overallSum  float64
		metrics     map[string]*metricAccumulator
	}

	teamsByDivision := make(map[string]*teamAccumulator)
	for _, member := range members {
		division := strings.TrimSpace(member.Division)
		if division == "" {
			division = "Tanpa Divisi"
		}
		team, ok := teamsByDivision[division]
		if !ok {
			team = &teamAccumulator{
				division: division,
				metrics:  make(map[string]*metricAccumulator, len(amaliyahMetricOrder)),
			}
			teamsByDivision[division] = team
		}

		team.memberCount++
		team.overallSum += member.OverallScore
		for _, metric := range member.Metrics {
			if metric.Available <= 0 {
				continue
			}
			acc, ok := team.metrics[metric.Key]
			if !ok {
				acc = &metricAccumulator{}
				team.metrics[metric.Key] = acc
			}
			acc.scoreSum += metric.Score
			acc.memberCount++
			acc.completedSum += metric.Completed
			acc.availableSum += metric.Available
		}
	}

	teams := make([]models.AmaliyahTeamLeaderboard, 0, len(teamsByDivision))
	for _, team := range teamsByDivision {
		item := models.AmaliyahTeamLeaderboard{
			Division:    team.division,
			MemberCount: team.memberCount,
			Metrics:     make([]models.AmaliyahMetricScore, 0, len(amaliyahMetricOrder)),
		}
		if team.memberCount > 0 {
			item.OverallScore = team.overallSum / float64(team.memberCount)
		}
		for _, key := range amaliyahMetricOrder {
			metric := models.AmaliyahMetricScore{Key: key, Label: amaliyahMetricLabels[key]}
			if acc := team.metrics[key]; acc != nil {
				metric.Completed = acc.completedSum
				metric.Available = acc.availableSum
				if acc.memberCount > 0 {
					metric.Score = acc.scoreSum / float64(acc.memberCount)
				}
			}
			item.Metrics = append(item.Metrics, metric)
		}
		teams = append(teams, item)
	}

	sort.SliceStable(teams, func(i, j int) bool {
		if teams[i].OverallScore == teams[j].OverallScore {
			return strings.ToLower(teams[i].Division) < strings.ToLower(teams[j].Division)
		}
		return teams[i].OverallScore > teams[j].OverallScore
	})
	for i := range teams {
		teams[i].Rank = i + 1
	}
	return teams
}
