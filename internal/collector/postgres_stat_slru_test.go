package collector

import (
	"database/sql"
	"testing"

	"github.com/cherts/pgscv/internal/model"
	"github.com/jackc/pgproto3/v2"
	"github.com/stretchr/testify/assert"
)

func TestPostgresStatSlruCollector_Update(t *testing.T) {
	var input = pipelineInput{
		required: []string{
			"postgres_stat_slru_blks_zeroed",
			"postgres_stat_slru_blks_hit",
			"postgres_stat_slru_blks_read",
			"postgres_stat_slru_blks_written",
			"postgres_stat_slru_blks_exists",
			"postgres_stat_slru_flushes",
			"postgres_stat_slru_truncates",
		},
		collector: NewPostgresStatSlruCollector,
		service:   model.ServiceTypePostgresql,
	}

	pipeline(t, input)
}

func Test_parsePostgresStatSlru(t *testing.T) {
	var testCases = []struct {
		name string
		res  *model.PGResult
		want map[string]postgresStatSlru
	}{
		{
			name: "normal output, Postgres 13",
			res: &model.PGResult{
				Nrows: 1,
				Ncols: 17,
				Colnames: []pgproto3.FieldDescription{
					{Name: []byte("name")}, {Name: []byte("blks_zeroed")}, {Name: []byte("blks_hit")}, {Name: []byte("blks_read")},
					{Name: []byte("blks_written")}, {Name: []byte("blks_exists")}, {Name: []byte("flushes")}, {Name: []byte("truncates")},
				},
				Rows: [][]sql.NullString{
					{
						{String: "subtransaction", Valid: true}, {String: "2972", Valid: true}, {String: "0", Valid: true}, {String: "0", Valid: true},
						{String: "2867", Valid: true}, {String: "0", Valid: true}, {String: "527", Valid: true}, {String: "527", Valid: true},
					},
				},
			},
			want: map[string]postgresStatSlru{
				"subtransaction": {
					SlruName: "subtransaction", BlksZeroed: 2972, BlksHit: 0, BlksRead: 0,
					BlksWritten: 2867, BlksExists: 0, Flushes: 527, Truncates: 527,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := parsePostgresStatSlru(tc.res, []string{"name"})
			assert.EqualValues(t, tc.want, got)
		})
	}
}
