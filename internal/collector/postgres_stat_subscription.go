package collector

import (
	"strconv"

	"github.com/cherts/pgscv/internal/log"
	"github.com/cherts/pgscv/internal/model"
	"github.com/cherts/pgscv/internal/store"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	postgresStatSubscriptionQuery14 = "SELECT pid, subname, COALESCE(relid::regclass::text, 'main') AS relname, " +
		"COALESCE(NULL::text, 'nothing') AS worker_type, " +
		"pg_wal_lsn_diff(pg_current_wal_lsn(), received_lsn) AS received, " +
		"pg_wal_lsn_diff(pg_current_wal_lsn(), latest_end_lsn) AS latest_end, " +
		"NULL::numeric AS apply_error_count, NULL::numeric AS sync_error_count " +
		"FROM pg_stat_subscription"

	postgresStatSubscriptionQuery16 = "SELECT s1.pid, s1.subname, COALESCE(s1.relid::regclass::text, 'main') AS relname, " +
		"COALESCE(NULL::text, 'nothing') AS worker_type, " +
		"pg_wal_lsn_diff(pg_current_wal_lsn(), s1.received_lsn) AS received, " +
		"pg_wal_lsn_diff(pg_current_wal_lsn(), s1.latest_end_lsn) AS latest_end, " +
		"s2.apply_error_count, s2.sync_error_count " +
		"FROM pg_stat_subscription s1 JOIN pg_stat_subscription_stats s2 ON s1.subid = s2.subid"

	postgresStatSubscriptionQueryLatest = "SELECT s1.pid, s1.subname, COALESCE(s1.relid::regclass::text, 'main') AS relname, " +
		"s1.worker_type AS worker_type, " +
		"pg_wal_lsn_diff(pg_current_wal_lsn(), s1.received_lsn) AS received, " +
		"pg_wal_lsn_diff(pg_current_wal_lsn(), s1.latest_end_lsn) AS latest_end, " +
		"s2.apply_error_count, s2.sync_error_count " +
		"FROM pg_stat_subscription s1 JOIN pg_stat_subscription_stats s2 ON s1.subid = s2.subid"
)

// postgresStatSubscriptionCollector defines metric descriptors and stats store.
type postgresStatSubscriptionCollector struct {
	labelNames []string
	lagBytes   typedDesc
	errorCount typedDesc
}

// NewPostgresStatSubscriptionCollector returns a new Collector exposing postgres pg_stat_subscription stats.
// For details see https://www.postgresql.org/docs/current/monitoring-stats.html#MONITORING-PG-STAT-SUBSCRIPTION
func NewPostgresStatSubscriptionCollector(constLabels labels, settings model.CollectorSettings) (Collector, error) {
	var labelNames = []string{"subname", "relname", "worker_type"}

	return &postgresStatSubscriptionCollector{
		labelNames: labelNames,
		lagBytes: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_subscription", "lag_bytes", "Number of bytes receiver is behind than sender in each WAL processing phase.", 0},
			prometheus.GaugeValue,
			labelNames, constLabels,
			settings.Filters,
		),
		errorCount: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_subscription", "error_count", "Number of times an error occurred.", 0},
			prometheus.GaugeValue,
			labelNames, constLabels,
			settings.Filters,
		),
	}, nil
}

// Update method collects statistics, parse it and produces metrics that are sent to Prometheus.
func (c *postgresStatSubscriptionCollector) Update(config Config, ch chan<- prometheus.Metric) error {
	if config.serverVersionNum < PostgresV10 {
		log.Debugln("[postgres stat_subscription collector]: pg_stat_subscription view are not available, required Postgres 10 or newer")
		return nil
	}

	conn, err := store.New(config.ConnString, config.ConnTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Collecting pg_stat_subscription since Postgres 10.
	if config.serverVersionNum >= PostgresV10 {
		res, err := conn.Query(selectSubscriptionQuery(config.serverVersionNum))
		if err != nil {
			log.Warnf("get pg_stat_subscription failed: %s; skip", err)
		} else {
			// Parse pg_stat_subscription stats.
			stats := parsePostgresSubscriptionStat(res, c.labelNames)
			for _, stat := range stats {
				if value, ok := stat.values["received"]; ok {
					ch <- c.lagBytes.newConstMetric(value, stat.SubName, stat.RelName, stat.WorkerType, "received")
				}
				if value, ok := stat.values["latest_end"]; ok {
					ch <- c.lagBytes.newConstMetric(value, stat.SubName, stat.RelName, stat.WorkerType, "latest_end")
				}
				if value, ok := stat.values["apply_error_count"]; ok {
					ch <- c.errorCount.newConstMetric(value, stat.SubName, stat.RelName, stat.WorkerType, "apply")
				}
				if value, ok := stat.values["sync_error_count"]; ok {
					ch <- c.errorCount.newConstMetric(value, stat.SubName, stat.RelName, stat.WorkerType, "sync")
				}
			}
		}
	}

	return nil
}

// postgresSubscriptionStat represents per-subscription stats based on pg_stat_subscription.
type postgresSubscriptionStat struct {
	Pid        string // a pid
	SubName    string // a subscription name
	RelName    string // a relname
	WorkerType string // a worker_type
	values     map[string]float64
}

// parsePostgresSubscriptionStat parses PGResult and returns structs with stats values.
func parsePostgresSubscriptionStat(r *model.PGResult, labelNames []string) map[string]postgresSubscriptionStat {
	log.Debug("parse postgres stat_subscription stats")

	var stats = make(map[string]postgresSubscriptionStat)

	for _, row := range r.Rows {
		stat := postgresSubscriptionStat{values: map[string]float64{}}

		// collect label values
		for i, colname := range r.Colnames {
			switch string(colname.Name) {
			case "pid":
				stat.Pid = row[i].String
			case "subname":
				stat.SubName = row[i].String
			case "relname":
				stat.RelName = row[i].String
			case "worker_type":
				stat.WorkerType = row[i].String
			}
		}

		// use pid as key in the map
		pid := stat.Pid

		// Put stats with labels (but with no data values yet) into stats store.
		stats[pid] = stat

		// fetch data values from columns
		for i, colname := range r.Colnames {
			// skip columns if its value used as a label
			if stringsContains(labelNames, string(colname.Name)) {
				continue
			}

			// Skip empty (NULL) values.
			if !row[i].Valid {
				continue
			}

			// Get data value and convert it to float64 used by Prometheus.
			v, err := strconv.ParseFloat(row[i].String, 64)
			if err != nil {
				log.Errorf("invalid input, parse '%s' failed: %s; skip", row[i].String, err)
				continue
			}

			s := stats[pid]

			// Run column-specific logic
			switch string(colname.Name) {
			case "received":
				s.values["received"] = v
			case "latest_end":
				s.values["latest_end"] = v
			case "apply_error_count":
				s.values["apply_error_count"] = v
			case "sync_error_count":
				s.values["sync_error_count"] = v
			default:
				continue
			}

			stats[pid] = s
		}
	}

	return stats
}

// selectSubscriptionQuery returns suitable subscription query depending on passed version.
func selectSubscriptionQuery(version int) string {
	switch {
	case version < PostgresV15:
		return postgresStatSubscriptionQuery14
	case version < PostgresV17:
		return postgresStatSubscriptionQuery16
	default:
		return postgresStatSubscriptionQueryLatest
	}
}
