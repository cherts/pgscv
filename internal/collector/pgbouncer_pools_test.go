package collector

import (
	"database/sql"
	"testing"

	"github.com/cherts/pgscv/internal/model"
	"github.com/jackc/pgproto3/v2"
	"github.com/stretchr/testify/assert"
)

func TestPgbouncerPoolsCollector_Update(t *testing.T) {
	var input = pipelineInput{
		required: []string{
			"pgbouncer_pool_connections_in_flight",
			"pgbouncer_pool_max_wait_seconds",
			"pgbouncer_client_connections_in_flight",
		},
		collector: NewPgbouncerPoolsCollector,
		service:   model.ServiceTypePgbouncer,
	}

	pipeline(t, input)
}

func Test_parsePgbouncerPoolsStats(t *testing.T) {
	var testCases = []struct {
		name string
		res  *model.PGResult
		want map[string]pgbouncerPoolStat
	}{
		{
			name: "normal output",
			res: &model.PGResult{
				Nrows: 2,
				Ncols: 15,
				Colnames: []pgproto3.FieldDescription{
					{Name: []byte("database")}, {Name: []byte("user")}, {Name: []byte("pool_mode")},
					{Name: []byte("cl_active")}, {Name: []byte("cl_waiting")}, {Name: []byte("cl_active_cancel_req")}, {Name: []byte("cl_waiting_cancel_req")},
					{Name: []byte("sv_active")}, {Name: []byte("sv_active_cancel")}, {Name: []byte("sv_being_canceled")}, {Name: []byte("sv_idle")},
					{Name: []byte("sv_used")}, {Name: []byte("sv_tested")}, {Name: []byte("sv_login")}, {Name: []byte("maxwait")},
				},
				Rows: [][]sql.NullString{
					{
						{String: "testdb1", Valid: true}, {String: "testuser1", Valid: true}, {String: "transaction", Valid: true},
						{String: "15", Valid: true}, {String: "5", Valid: true}, {String: "0", Valid: true}, {String: "0", Valid: true},
						{String: "10", Valid: true}, {String: "0", Valid: true}, {String: "0", Valid: true}, {String: "1", Valid: true},
						{String: "1", Valid: true}, {String: "1", Valid: true}, {String: "1", Valid: true}, {String: "1", Valid: true},
					},
					{
						{String: "testdb2", Valid: true}, {String: "testuser2", Valid: true}, {String: "statement", Valid: true},
						{String: "25", Valid: true}, {String: "10", Valid: true}, {String: "0", Valid: true}, {String: "5", Valid: true},
						{String: "25", Valid: true}, {String: "2", Valid: true}, {String: "0", Valid: true}, {String: "1", Valid: true},
						{String: "2", Valid: true}, {String: "2", Valid: true}, {String: "2", Valid: true}, {String: "2", Valid: true},
					},
				},
			},
			want: map[string]pgbouncerPoolStat{
				"testuser1/testdb1/transaction": {
					database: "testdb1", user: "testuser1", mode: "transaction",
					clActive: 15, clWaiting: 5, clActiveCancelReq: 0, clWaitingCancelReq: 0,
					svActive: 10, svActiveCancel: 0, svBeingCanceled: 0, svIdle: 1,
					svUsed: 1, svTested: 1, svLogin: 1, maxWait: 1,
				},
				"testuser2/testdb2/statement": {
					database: "testdb2", user: "testuser2", mode: "statement",
					clActive: 25, clWaiting: 10, clActiveCancelReq: 0, clWaitingCancelReq: 5,
					svActive: 25, svActiveCancel: 2, svBeingCanceled: 0, svIdle: 1,
					svUsed: 2, svTested: 2, svLogin: 2, maxWait: 2,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := parsePgbouncerPoolsStats(tc.res, []string{"database", "user", "pool_mode"})
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func Test_parsePgbouncerClientsStats(t *testing.T) {
	var testCases = []struct {
		name string
		res  *model.PGResult
		want map[string]float64
	}{
		{
			name: "normal output",
			res: &model.PGResult{
				Nrows: 10,
				Ncols: 5,
				Colnames: []pgproto3.FieldDescription{
					{Name: []byte("user")}, {Name: []byte("database")}, {Name: []byte("addr")}, {Name: []byte("state")}, {Name: []byte("port")},
				},
				Rows: [][]sql.NullString{
					{{String: "user1", Valid: true}, {String: "db1", Valid: true}, {String: "1.1.1.1", Valid: true}, {String: "active", Valid: true}, {String: "11", Valid: true}},
					{{String: "user2", Valid: true}, {String: "db2", Valid: true}, {String: "2.2.2.2", Valid: true}, {String: "idle", Valid: true}, {String: "22", Valid: true}},
					{{String: "user1", Valid: true}, {String: "db1", Valid: true}, {String: "1.1.1.1", Valid: true}, {String: "active", Valid: true}, {String: "12", Valid: true}},
					{{String: "user3", Valid: true}, {String: "db3", Valid: true}, {String: "unix", Valid: true}, {String: "active", Valid: true}, {String: "unix", Valid: true}},
					{{String: "user3", Valid: true}, {String: "db3", Valid: true}, {String: "unix", Valid: true}, {String: "idle", Valid: true}, {String: "unix", Valid: true}},
					{{String: "user2", Valid: true}, {String: "db2", Valid: true}, {String: "2.2.2.2", Valid: true}, {String: "active", Valid: true}, {String: "23", Valid: true}},
					{{String: "user1", Valid: true}, {String: "db1", Valid: true}, {String: "1.1.1.1", Valid: true}, {String: "active", Valid: true}, {String: "13", Valid: true}},
					{{String: "user1", Valid: true}, {String: "db1", Valid: true}, {String: "1.1.1.1", Valid: true}, {String: "idle", Valid: true}, {String: "14", Valid: true}},
					{{String: "user2", Valid: true}, {String: "db2", Valid: true}, {String: "2.2.2.2", Valid: true}, {String: "active", Valid: true}, {String: "24", Valid: true}},
					{{String: "user1", Valid: true}, {String: "db1", Valid: true}, {String: "1.1.1.1", Valid: true}, {String: "active", Valid: true}, {String: "25", Valid: true}},
				},
			},
			want: map[string]float64{
				"user1/db1/1.1.1.1": 5,
				"user2/db2/2.2.2.2": 3,
				"user3/db3/unix":    2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := parsePgbouncerClientsStats(tc.res)
			assert.EqualValues(t, tc.want, got)
		})
	}
}
