package services

import (
	"testing"

	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/stretchr/testify/require"
)

func TestBuildTeamLeaderboardUsesAverageMemberPercentages(t *testing.T) {
	members := []models.AmaliyahMemberSnapshot{
		{
			MemberID:     1,
			Division:     "Admin Kelas",
			OverallScore: 80,
			Metrics: []models.AmaliyahMetricScore{
				{Key: models.AmaliyahTahajud, Available: 30, Completed: 24, Score: 80},
				{Key: models.AmaliyahTilawahTadabur, Available: 30, Completed: 27, Score: 90},
			},
		},
		{
			MemberID:     2,
			Division:     "Admin Kelas",
			OverallScore: 100,
			Metrics: []models.AmaliyahMetricScore{
				{Key: models.AmaliyahTahajud, Available: 10, Completed: 10, Score: 100},
				{Key: models.AmaliyahTilawahTadabur, Available: 10, Completed: 10, Score: 100},
			},
		},
		{
			MemberID:     3,
			Division:     "Marketing",
			OverallScore: 85,
			Metrics: []models.AmaliyahMetricScore{
				{Key: models.AmaliyahTahajud, Available: 30, Completed: 25, Score: 83.3333},
			},
		},
	}

	teams := buildTeamLeaderboard(members)
	require.Len(t, teams, 2)
	require.Equal(t, "Admin Kelas", teams[0].Division)
	require.Equal(t, 1, teams[0].Rank)
	require.Equal(t, 2, teams[0].MemberCount)
	require.Equal(t, float64(90), teams[0].OverallScore)

	metricByKey := map[string]models.AmaliyahMetricScore{}
	for _, metric := range teams[0].Metrics {
		metricByKey[metric.Key] = metric
	}
	require.Equal(t, float64(90), metricByKey[models.AmaliyahTahajud].Score)
	require.Equal(t, float64(95), metricByKey[models.AmaliyahTilawahTadabur].Score)
}
