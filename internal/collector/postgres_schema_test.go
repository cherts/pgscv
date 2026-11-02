package collector

import (
	"context"
	"github.com/cherts/pgscv/internal/model"
	"github.com/cherts/pgscv/internal/store"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPostgresSchemaCollector_Update(t *testing.T) {
	var input = pipelineInput{
		required: []string{
			"postgres_schema_system_catalog_bytes",
			"postgres_schema_non_pk_tables",
			"postgres_schema_invalid_indexes_bytes",
			"postgres_schema_non_indexed_fkeys",
			"postgres_schema_redundant_indexes_bytes",
			"postgres_schema_sequence_exhaustion_ratio",
			"postgres_schema_mistyped_fkeys",
		},
		collector: NewPostgresSchemasCollector,
		service:   model.ServiceTypePostgresql,
	}

	pipeline(t, input)
}

func Test_getSystemCatalogSize(t *testing.T) {
	conn := store.NewTest(t)
	got, err := getSystemCatalogSize(conn)
	assert.NoError(t, err)
	assert.NotEqual(t, float64(0), got)

	_ = conn.Conn().Close(context.Background())
	got, err = getSystemCatalogSize(conn)
	assert.Error(t, err)
	assert.Equal(t, float64(0), got)
}

func Test_getSchemaNonPKTables(t *testing.T) {
	conn := store.NewTest(t)
	got, err := getSchemaNonPKTables(conn)
	assert.NoError(t, err)
	assert.Less(t, 0, len(got))

	_ = conn.Conn().Close(context.Background())
	got, err = getSchemaNonPKTables(conn)
	assert.Error(t, err)
	assert.Equal(t, 0, len(got))
}

func Test_getSchemaInvalidIndexes(t *testing.T) {
	conn := store.NewTest(t)

	var serverVersion int
	err := conn.Conn().QueryRow(context.Background(),
		"SELECT setting::int FROM pg_catalog.pg_settings WHERE name = 'server_version_num'").Scan(&serverVersion)
	assert.NoError(t, err)

	// Query with pg_stat_progress_create_index could be tested on Postgres 12 and newer only.
	versions := []int{PostgresV95}
	if serverVersion >= PostgresV12 {
		versions = append(versions, PostgresV12)
	}

	// Both query variants should find fixture's invalid index.
	for _, version := range versions {
		got, err := getSchemaInvalidIndexes(conn, version)
		assert.NoError(t, err)
		assert.Less(t, 0, len(got))
	}

	_ = conn.Conn().Close(context.Background())
	got, err := getSchemaInvalidIndexes(conn, PostgresV12)
	assert.Error(t, err)
	assert.Equal(t, 0, len(got))
}

func Test_selectSchemaInvalidIndexesQuery(t *testing.T) {
	testcases := []struct {
		version int
		want    string
	}{
		{version: PostgresV95, want: schemaInvalidIndexesQuery11},
		{version: PostgresV11, want: schemaInvalidIndexesQuery11},
		{version: PostgresV12, want: schemaInvalidIndexesQuery12},
		{version: PostgresV18, want: schemaInvalidIndexesQuery12},
	}

	for _, tc := range testcases {
		assert.Equal(t, tc.want, selectSchemaInvalidIndexesQuery(tc.version))
	}
}

func Test_getSchemaNonIndexedFK(t *testing.T) {
	conn := store.NewTest(t)
	got, err := getSchemaNonIndexedFK(conn)
	assert.NoError(t, err)
	assert.Less(t, 0, len(got))

	_ = conn.Conn().Close(context.Background())
	got, err = getSchemaNonIndexedFK(conn)
	assert.Error(t, err)
	assert.Equal(t, 0, len(got))
}

func Test_getSchemaRedundantIndexes(t *testing.T) {
	conn := store.NewTest(t)
	got, err := getSchemaRedundantIndexes(conn)
	assert.NoError(t, err)
	assert.Less(t, 0, len(got))

	_ = conn.Conn().Close(context.Background())
	got, err = getSchemaRedundantIndexes(conn)
	assert.Error(t, err)
	assert.Equal(t, 0, len(got))
}

func Test_getSchemaSequences(t *testing.T) {
	conn := store.NewTest(t)
	got, err := getSchemaSequences(conn)
	assert.NoError(t, err)
	assert.Less(t, 0, len(got))

	_ = conn.Conn().Close(context.Background())
	got, err = getSchemaSequences(conn)
	assert.Error(t, err)
	assert.Equal(t, 0, len(got))
}

func Test_getSchemaFKDatatypeMismatch(t *testing.T) {
	conn := store.NewTest(t)
	got, err := getSchemaFKDatatypeMismatch(conn)
	assert.NoError(t, err)
	assert.Less(t, 0, len(got))

	_ = conn.Conn().Close(context.Background())
	got, err = getSchemaFKDatatypeMismatch(conn)
	assert.Error(t, err)
	assert.Equal(t, 0, len(got))
}
