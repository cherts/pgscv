package collector

import (
	"testing"

	"github.com/cherts/pgscv/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestPostgresSubscriptionRelCollector_Update(t *testing.T) {
	var input = pipelineInput{
		required: []string{
			"postgres_subscription_rel_count",
		},
		collector: NewPostgresSubscriptionRelCollector,
		service:   model.ServiceTypePostgresql,
	}

	pipelineLogicalReplication(t, input)
}

func Test_selectSubscriptionRelQuery(t *testing.T) {
	var testcases = []struct {
		version int
		want    string
	}{
		{version: 100000, want: postgresSubscriptionRel15},
		{version: 100005, want: postgresSubscriptionRel15},
		{version: 130002, want: postgresSubscriptionRel15},
		{version: 140005, want: postgresSubscriptionRel15},
		{version: 150001, want: postgresSubscriptionRel15},
		{version: 160002, want: postgresSubscriptionRelLatest},
		{version: 170005, want: postgresSubscriptionRelLatest},
		{version: 180000, want: postgresSubscriptionRelLatest},
	}

	for _, tc := range testcases {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tc.want, selectSubscriptionRelQuery(tc.version))
		})
	}
}
