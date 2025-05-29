package collector

import (
	"database/sql"
	"testing"

	"github.com/cherts/pgscv/internal/model"
	"github.com/jackc/pgproto3/v2"
	"github.com/stretchr/testify/assert"
)

func TestPostgresStatSslCollector_Update(t *testing.T) {
	var input = pipelineInput{
		required: []string{
			"postgres_stat_ssl_conn_number",
		},
		collector: NewPostgresStatSslCollector,
		service:   model.ServiceTypePostgresql,
	}

	pipeline(t, input)
}

func Test_parsePostgresStatSsl(t *testing.T) {
	var testCases = []struct {
		name string
		res  *model.PGResult
		want map[string]postgresStatSsl
	}{
		{
			name: "normal output, Postgres 16",
			res: &model.PGResult{
				Nrows: 1,
				Ncols: 3,
				Colnames: []pgproto3.FieldDescription{
					{Name: []byte("database")}, {Name: []byte("username")}, {Name: []byte("ssl_conn_number")},
				},
				Rows: [][]sql.NullString{
					{
						{String: "NULL", Valid: true}, {String: "postgres", Valid: true}, {String: "1", Valid: true},
					},
				},
			},
			want: map[string]postgresStatSsl{
				"NULL/postgres": {
					Database: "NULL", Username: "postgres", ConnNumber: 1,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := parsePostgresStatSsl(tc.res, []string{"database", "username"})
			assert.EqualValues(t, tc.want, got)
		})
	}
}
