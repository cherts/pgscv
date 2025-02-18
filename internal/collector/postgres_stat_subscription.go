package collector

import (
	"strconv"
	"strings"

	"github.com/cherts/pgscv/internal/log"
	"github.com/cherts/pgscv/internal/model"
	"github.com/cherts/pgscv/internal/store"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	postgresStatSubscriptionQuery14 = "SELECT subname, pid, COALESCE(relid::regclass, 'main') AS relname, " +
		"pg_wal_lsn_diff(pg_current_wal_lsn(), received_lsn) AS received, " +
		"pg_wal_lsn_diff(pg_current_wal_lsn(), latest_end_lsn) AS latest, " +
		"NULL::numeric AS apply_error_count, NULL::numeric AS sync_error_count " +
		"FROM pg_stat_subscription"

	postgresStatSubscriptionQueryLatest = "SELECT s1.subname, s1.pid, COALESCE(s1.relid::regclass, 'main') AS relname, " +
		"pg_wal_lsn_diff(pg_current_wal_lsn(), s1.received_lsn) AS received, " +
		"pg_wal_lsn_diff(pg_current_wal_lsn(), s1.latest_end_lsn) AS latest, " +
		"s2.apply_error_count, s2.sync_error_count " +
		"FROM pg_stat_subscription s1 JOIN pg_stat_subscription_stats s2 ON s1.subid = s2.subid"
)

// postgresStatSubscriptionCollector defines metric descriptors and stats store.
type postgresStatSubscriptionCollector struct {
	pid             typedDesc
	received        typedDesc
	latest          typedDesc
	applyErrorCount typedDesc
	syncErrorCount  typedDesc
	labelNames      []string
}

// NewPostgresStatSubscriptionCollector returns a new Collector exposing postgres pg_stat_subscription stats.
func NewPostgresStatSubscriptionCollector(constLabels labels, settings model.CollectorSettings) (Collector, error) {
	var labels = []string{"subname"}

	return &postgresStatSubscriptionCollector{
		pid: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_subscription", "pid", "XXXX", 0},
			prometheus.GaugeValue,
			labels, constLabels,
			settings.Filters,
		),
		received: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_subscription", "received", "XXXX", 0},
			prometheus.GaugeValue,
			labels, constLabels,
			settings.Filters,
		),
		latest: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_subscription", "latest", "XXXX", 0},
			prometheus.GaugeValue,
			labels, constLabels,
			settings.Filters,
		),
		applyErrorCount: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_subscription", "apply_error_count", "XXXX", 0},
			prometheus.GaugeValue,
			labels, constLabels,
			settings.Filters,
		),
		syncErrorCount: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_subscription", "sync_error_count", "XXXX", 0},
			prometheus.GaugeValue,
			labels, constLabels,
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
			stats := parsePostgresStatSubscription(res, []string{"subname"})

			for _, stat := range stats {
				ch <- c.pid.newConstMetric(stat.Pid, stat.SubName)
				ch <- c.received.newConstMetric(stat.Received, stat.SubName)
				ch <- c.latest.newConstMetric(stat.Latest, stat.SubName)
				ch <- c.applyErrorCount.newConstMetric(stat.ApplyErrorCount, stat.SubName)
				ch <- c.syncErrorCount.newConstMetric(stat.SyncErrorCount, stat.SubName)
			}
		}
	}

	return nil
}

// postgresStatSubscription
type postgresStatSubscription struct {
	SubName         string // a subscription name
	Pid             float64
	Received        float64
	Latest          float64
	ApplyErrorCount float64
	SyncErrorCount  float64
}

// parsePostgresStatSubscription parses PGResult and returns structs with stats values.
func parsePostgresStatSubscription(r *model.PGResult, labelNames []string) map[string]postgresStatSubscription {
	log.Debug("parse postgres stat_subscription stats")

	var stats = make(map[string]postgresStatSubscription)

	for _, row := range r.Rows {
		var SubName string

		for i, colname := range r.Colnames {
			switch string(colname.Name) {
			case "subname":
				SubName = row[i].String
			}
		}

		// create a stat_io name consisting of trio BackendType/IoObject/IoContext
		statSubsciption := strings.Join([]string{SubName}, "")

		// Put stats with labels (but with no data values yet) into stats store.
		if _, ok := stats[statSubsciption]; !ok {
			stats[statSubsciption] = postgresStatSubscription{SubName: SubName}
		}

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

			s := stats[statSubsciption]

			switch string(colname.Name) {
			case "pid":
				s.Pid = v
			case "received":
				s.Received = v
			case "latest":
				s.Latest = v
			case "apply_error_count":
				s.ApplyErrorCount = v
			case "sync_error_count":
				s.SyncErrorCount = v
			default:
				continue
			}

			stats[statSubsciption] = s
		}
	}

	return stats
}

// selectSubscriptionQuery returns suitable subscription query depending on passed version.
func selectSubscriptionQuery(version int) string {
	switch {
	case version < PostgresV15:
		return postgresStatSubscriptionQuery14
	default:
		return postgresStatSubscriptionQueryLatest
	}
}
