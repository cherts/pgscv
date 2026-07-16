package collector

import (
	"fmt"
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
		{version: PostgresV10, want: fmt.Sprintf(postgresSubscriptionRel15, "pgscv_fixtures")},
		{version: 100005, want: fmt.Sprintf(postgresSubscriptionRel15, "pgscv_fixtures")},
		{version: 130002, want: fmt.Sprintf(postgresSubscriptionRel15, "pgscv_fixtures")},
		{version: 140005, want: fmt.Sprintf(postgresSubscriptionRel15, "pgscv_fixtures")},
		{version: 150001, want: fmt.Sprintf(postgresSubscriptionRel15, "pgscv_fixtures")},
		{version: 160002, want: fmt.Sprintf(postgresSubscriptionRelLatest, "pgscv_fixtures")},
		{version: 170005, want: fmt.Sprintf(postgresSubscriptionRelLatest, "pgscv_fixtures")},
		{version: PostgresV18, want: fmt.Sprintf(postgresSubscriptionRelLatest, "pgscv_fixtures")},
	}

	for _, tc := range testcases {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tc.want, selectSubscriptionRelQuery(tc.version, "pgscv_fixtures"))
		})
	}
}
