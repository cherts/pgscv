package collector

import (
	"github.com/cherts/pgscv/internal/model"
	"testing"
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
