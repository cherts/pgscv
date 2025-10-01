package collector

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cherts/pgscv/internal/model"
	"github.com/cherts/pgscv/internal/store"
	"github.com/stretchr/testify/assert"
)

func Test_runTailLoop(t *testing.T) {
	c, err := NewPostgresLogsCollector(nil, model.CollectorSettings{})
	assert.NoError(t, err)
	assert.NotNil(t, c)
	lc := c.(*postgresLogsCollector)
	lc.logDestination = postgresLogsDestinationStderr

	fname1 := "/tmp/pgscv_postgres_logs_test_sample_1.log"
	fstr1 := "2020-09-30 14:26:29.777 +05 797922 LOG: PID 0 in cancel request did not match any process\n"
	fname2 := "/tmp/pgscv_postgres_logs_test_sample_2.log"
	fstr2 := "2020-09-30 14:26:29.784 +05 797923 ERROR: syntax error\n"

	// create test files
	for _, name := range []string{fname1, fname2} {
		f, err := os.Create(name)
		assert.NoError(t, err)
		assert.NoError(t, f.Close())
	}

	// tail first file from the beginning
	lc.updateLogfile <- fname1
	time.Sleep(200 * time.Millisecond)

	// write line to first file
	f, err := os.OpenFile(fname1, os.O_RDWR|os.O_APPEND, 0644)
	assert.NoError(t, err)
	_, err = f.WriteString(fstr1)
	assert.NoError(t, err)
	assert.NoError(t, f.Sync())
	assert.NoError(t, f.Close())
	time.Sleep(200 * time.Millisecond)

	// check store content -- should be log:1, error:0
	lc.totals.mu.RLock()
	assert.Equal(t, float64(1), lc.totals.store["log"])
	assert.Equal(t, float64(0), lc.totals.store["error"])
	lc.totals.mu.RUnlock()

	// tail second file
	lc.updateLogfile <- fname2
	time.Sleep(200 * time.Millisecond)

	// write line to second file
	f, err = os.OpenFile(fname2, os.O_RDWR|os.O_APPEND, 0644)
	assert.NoError(t, err)
	_, err = f.WriteString(fstr2)
	assert.NoError(t, err)
	assert.NoError(t, f.Sync())
	assert.NoError(t, f.Close())
	time.Sleep(200 * time.Millisecond)

	// check store content -- should be log:1, error:1
	lc.totals.mu.RLock()
	assert.Equal(t, float64(1), lc.totals.store["log"])
	assert.Equal(t, float64(1), lc.totals.store["error"])
	lc.totals.mu.RUnlock()

	// tail first file again (tail will start from the beginning, read all existing lines)
	lc.updateLogfile <- fname1
	time.Sleep(200 * time.Millisecond)

	// append one more line to first file
	f, err = os.OpenFile(fname1, os.O_RDWR|os.O_APPEND, 0644)
	assert.NoError(t, err)
	_, err = f.WriteString(fstr1)
	assert.NoError(t, err)
	assert.NoError(t, f.Sync())
	assert.NoError(t, f.Close())
	time.Sleep(200 * time.Millisecond)

	// check store content -- should be log:3, error:1 (3 because, 1 line in first reading and 2 lines in second reading.)
	lc.totals.mu.RLock()
	assert.Equal(t, float64(3), lc.totals.store["log"])
	assert.Equal(t, float64(1), lc.totals.store["error"])
	lc.totals.mu.RUnlock()

	// truncate second file.
	f, err = os.OpenFile(fname2, os.O_RDWR|os.O_TRUNC, 0644)
	assert.NoError(t, err)
	assert.NoError(t, f.Sync())
	assert.NoError(t, f.Close())

	// tail second (truncated) file again
	lc.updateLogfile <- fname2
	time.Sleep(200 * time.Millisecond)

	// add line to second file
	f, err = os.OpenFile(fname2, os.O_RDWR|os.O_APPEND, 0644)
	assert.NoError(t, err)
	_, err = f.WriteString(fstr2)
	assert.NoError(t, err)
	assert.NoError(t, f.Sync())
	assert.NoError(t, f.Close())
	time.Sleep(200 * time.Millisecond)

	// check store content -- should be log:3, error:2
	lc.totals.mu.RLock()
	assert.Equal(t, float64(3), lc.totals.store["log"])
	assert.Equal(t, float64(2), lc.totals.store["error"])
	lc.totals.mu.RUnlock()

	// remove test files
	for _, name := range []string{fname1, fname2} {
		assert.NoError(t, os.Remove(name))
	}
}

func Test_tailCollect(t *testing.T) {
	c, err := NewPostgresLogsCollector(nil, model.CollectorSettings{})
	assert.NoError(t, err)
	assert.NotNil(t, c)
	lc := c.(*postgresLogsCollector)
	lc.logDestination = postgresLogsDestinationStderr

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	tailCollect(ctx, "testdata/datadir/postgresql.log.golden", false, &wg, lc)
	assert.Equal(t, float64(6), lc.totals.store["log"])
	assert.Equal(t, float64(1), lc.totals.store["error"])
	assert.Equal(t, float64(2), lc.totals.store["fatal"])

	wg.Wait()
}

func Test_tailCollect_jsonLogParser(t *testing.T) {
	c, err := NewPostgresLogsCollector(nil, model.CollectorSettings{})
	assert.NoError(t, err)
	assert.NotNil(t, c)
	lc := c.(*postgresLogsCollector)
	lc.logDestination = postgresLogsDestinationJSON

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	tailCollect(ctx, "testdata/datadir/postgresql.jsonlog.golden", false, &wg, lc)
	assert.Equal(t, float64(2), lc.totals.store["log"])
	assert.Equal(t, float64(3), lc.totals.store["error"])
	assert.Equal(t, float64(0), lc.totals.store["fatal"])

	wg.Wait()
}

func Test_queryCurrentLogfile(t *testing.T) {
	got, err := queryCurrentLogfile(Config{DB: store.NewTest(t)})
	assert.NoError(t, err)
	assert.NotEqual(t, got, "")

	got, err = queryCurrentLogfile(Config{DB: nil})
	assert.Error(t, err)
	assert.Equal(t, got, "")

	got, err = queryCurrentLogfile(Config{DB: store.NewTest(t), LogDirectory: "/custom/dir"})
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(got, "/custom/dir"))
}

func Test_newStderrLogParser(t *testing.T) {
	p := newStderrLogParser()
	assert.NotNil(t, p)
	assert.Greater(t, len(p.reSeverity), 0)
	assert.Greater(t, len(p.reNormalize), 0)
}

func Test_stderrLogParser_updateMessagesStats(t *testing.T) {
	c, err := NewPostgresLogsCollector(nil, model.CollectorSettings{})
	assert.NoError(t, err)
	assert.NotNil(t, c)
	lc := c.(*postgresLogsCollector)

	p := newStderrLogParser()

	f, err := os.Open("testdata/datadir/postgresql.log.golden")
	assert.NoError(t, err)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		p.updateMessagesStats(scanner.Text(), lc)
	}

	lc.totals.mu.RLock()
	//fmt.Println(lc.totals.store)
	assert.Equal(t, float64(2), lc.totals.store["fatal"])
	assert.Equal(t, float64(1), lc.totals.store["error"])
	assert.Equal(t, float64(6), lc.totals.store["log"])
	lc.totals.mu.RUnlock()

	lc.fatals.mu.RLock()
	fmt.Println()
	assert.Equal(t, 1, len(lc.fatals.store))
	lc.fatals.mu.RUnlock()

	lc.panics.mu.RLock()
	assert.Equal(t, 0, len(lc.panics.store))
	lc.panics.mu.RUnlock()
}

func Test_stderrLogParser_parseMessageSeverity(t *testing.T) {
	testcases := []struct {
		line  string
		want  string
		found bool
	}{
		{line: "2020-09-29 14:08:52.858 +05 1060 [] LOG: test", want: "log", found: true},
		{line: "2020-09-29 14:08:52.858 +05 1060 []LOG: test", want: "log", found: true},
		{line: "2020-09-29 14:08:52.858 +05 1060 [] WARNING: test", want: "warning", found: true},
		{line: "2020-09-29 14:08:52.858 +05 1060 [] WARNING:  test", want: "warning", found: true},
		{line: "2020-09-29 14:08:52.858 +05 1060 [] ERROR: test", want: "error", found: true},
		{line: "2020-09-29 14:08:52.858 +05 1060 [] FATAL: test", want: "fatal", found: true},
		{line: "2020-09-29 14:08:52.858 +05 1060 [] PANIC: test", want: "panic", found: true},
		{line: "2020-09-29 14:08:52.858 +05 1060 [] STATEMENT: select log:(test)", want: "", found: false},
		{line: "2020-09-29 14:08:52.858 +05 1060 [] STATEMENT: select fn(WARNING:)", want: "", found: false},
		{line: "2020-09-29 14:08:52.858 +05 1060 [] STATEMENT: select error, test", want: "", found: false},
		{line: "", want: "", found: false},
		{line: "test", want: "", found: false},
	}

	p := newStderrLogParser()

	for _, tc := range testcases {
		got, ok := p.parseMessageSeverity(tc.line)
		assert.Equal(t, tc.want, got)
		assert.Equal(t, tc.found, ok)
	}
}

func Test_stderrLogParser_normalizeMessage(t *testing.T) {
	testcases := []struct {
		in   string
		want string
	}{
		{
			in:   `2020-10-01 08:37:58.208 +05 1402271 PANIC:  syntax error at or near "invalid" at character 1`,
			want: `syntax error at or near ? at character ?`,
		},
		{
			in:   `2020-10-01 08:37:58.208 +05 1402271 ERROR:  syntax error at or near ")" at character 721`,
			want: `syntax error at or near ? at character ?`,
		},
		{
			in:   `2020-10-01 08:37:58.208 +05 1402271 ERROR:  duplicate key value violates unique constraint "oauth_access_token_authentication_id_uindex"`,
			want: `duplicate key value violates unique constraint ?`,
		},
		{
			in:   `2020-10-01 08:37:58.208 +05 1402271 ERROR:  insert or update on table "offer_offer_products" violates foreign key constraint "fkbbwwdruntj50nng01y0g98cof"`,
			want: `insert or update on table ? violates foreign key constraint ?`,
		},
		{
			in:   `2020-10-01 08:37:58.208 +05 1402271 ERROR:  could not serialize access due to concurrent update`,
			want: `could not serialize access due to concurrent update`,
		},
		{
			in:   `2020-10-01 08:37:58.208 +05 1402271 LOG:  test`,
			want: "",
		},
	}

	parser := newStderrLogParser()

	for _, tc := range testcases {
		assert.Equal(t, tc.want, parser.normalizeMessage(tc.in))
	}
}

func Test_jsonLogParser_normalizeMessage(t *testing.T) {
	testcases := []struct {
		jsonLine jsonLine
		want     string
	}{
		{
			jsonLine: jsonLine{
				Message:       `syntax error at or near "invalid" at character 1`,
				ErrorSeverity: "ERROR",
			},
			want: `syntax error at or near ? at character ?`,
		},
		{
			jsonLine: jsonLine{
				Message:       `syntax error at or near ")" at character 721`,
				ErrorSeverity: "ERROR",
			},
			want: `syntax error at or near ? at character ?`,
		},
		{
			jsonLine: jsonLine{
				Message:       `duplicate key value violates unique constraint "oauth_access_token_authentication_id_uindex"`,
				ErrorSeverity: "ERROR",
			},
			want: `duplicate key value violates unique constraint ?`,
		},
		{
			jsonLine: jsonLine{
				Message:       `insert or update on table "offer_offer_products" violates foreign key constraint "fkbbwwdruntj50nng01y0g98cof"`,
				ErrorSeverity: "ERROR",
			},
			want: `insert or update on table ? violates foreign key constraint ?`,
		},
		{
			jsonLine: jsonLine{
				Message:       `could not serialize access due to concurrent update`,
				ErrorSeverity: "ERROR",
			},
			want: `could not serialize access due to concurrent update`,
		},
		{
			jsonLine: jsonLine{
				Message:       `test`,
				ErrorSeverity: "LOG",
			},
			want: "",
		},
	}

	parser := newJSONLogParser()

	for _, tc := range testcases {
		assert.Equal(t, tc.want, parser.normalizeMessage(tc.jsonLine))
	}
}

func Test_jsonLogParser_jsonLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    jsonLine
		wantErr bool
	}{
		{
			line: `{"error_severity":"ERROR","message":"syntax error at or near \"invalid\" at character 1"}`,
			want: jsonLine{
				ErrorSeverity: "ERROR",
				Message:       "syntax error at or near \"invalid\" at character 1",
			},
		},
		{
			line: `{"error_severity":"ERROR","message":"duplicate key value violates unique constraint \"oauth_access_token_authentication_id_uindex\""}`,
			want: jsonLine{
				ErrorSeverity: "ERROR",
				Message:       "duplicate key value violates unique constraint \"oauth_access_token_authentication_id_uindex\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := newJSONLogParser()

			jsonLine, err := parser.jsonLine(tt.line)
			if tt.wantErr {
				assert.Error(t, err)

				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, jsonLine)
		})
	}
}

func Test_jsonLogParser_updateMessagesStats(t *testing.T) {
	c, err := NewPostgresLogsCollector(nil, model.CollectorSettings{})
	assert.NoError(t, err)
	assert.NotNil(t, c)
	lc := c.(*postgresLogsCollector)

	p := newJSONLogParser()

	f, err := os.Open("testdata/datadir/postgresql.jsonlog.golden")
	assert.NoError(t, err)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		p.updateMessagesStats(scanner.Text(), lc)
	}

	lc.totals.mu.RLock()
	assert.Equal(t, float64(3), lc.totals.store["error"])
	assert.Equal(t, float64(2), lc.totals.store["log"])
	lc.totals.mu.RUnlock()

	lc.fatals.mu.RLock()
	fmt.Println()
	assert.Equal(t, 1, len(lc.errors.store))
	lc.fatals.mu.RUnlock()

	lc.panics.mu.RLock()
	assert.Equal(t, 0, len(lc.panics.store))
	lc.panics.mu.RUnlock()
}
