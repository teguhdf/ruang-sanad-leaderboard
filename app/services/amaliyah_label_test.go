package services

import (
	"testing"

	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/stretchr/testify/require"
)

func TestClassifyAmaliyahMetricLabelVariants(t *testing.T) {
	tests := map[string]string{
		"Tahajjud":             models.AmaliyahTahajud,
		"Tilawah":              models.AmaliyahTilawahTadabur,
		"Tilawah + Tadabur":    models.AmaliyahTilawahTadabur,
		"Dzikir Pagi":          models.AmaliyahDzikirPagi,
		"Zikir Pagi":           models.AmaliyahDzikirPagi,
		"Dzikir Petang":        models.AmaliyahDzikirPetang,
		"Dzikir Sore":          models.AmaliyahDzikirPetang,
		"Zikir Petang / Sore":  models.AmaliyahDzikirPetang,
		"Dzikir Malam":         models.AmaliyahDzikirPetang,
	}

	for label, want := range tests {
		t.Run(label, func(t *testing.T) {
			require.Equal(t, want, classifyAmaliyahMetricLabel(label))
		})
	}
}
