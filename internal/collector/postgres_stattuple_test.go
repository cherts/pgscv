package collector

import (
	"github.com/cherts/pgscv/internal/model"
	"testing"
)

func TestPostgresSubscriptionRelCollector_Update(t *testing.T) {
	var input = pipelineInput{
		required: []string{
			"postgres_pgstattuple_approx_free_percent",
			"postgres_pgstattuple_approx_free_space",
			"postgres_pgstattuple_approx_tuple_percent",
			"postgres_pgstattuple_dead_tuple_count",
			"postgres_pgstattuple_dead_tuple_len",
			"postgres_pgstattuple_dead_tuple_percent",
		},
		collector: NewPostgresStatTupleCollector,
		service:   model.ServiceTypePostgresql,
	}

	pipeline(t, input)
}
