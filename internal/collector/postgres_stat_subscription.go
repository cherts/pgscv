package collector

import (
	"strconv"

	"github.com/cherts/pgscv/internal/log"
	"github.com/cherts/pgscv/internal/model"
	"github.com/cherts/pgscv/internal/store"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	postgresStatSubscriptionQuery14 = "SELECT subid, subname, COALESCE(s1.pid, 0) AS pid, " +
		"COALESCE(NULL::text, 'unknown') AS worker_type, " +
		"COALESCE(received_lsn - '0/0', 0) AS received_lsn, " +
		"COALESCE(latest_end_lsn - '0/0', 0) AS reported_lsn, " +
		"COALESCE(EXTRACT(EPOCH FROM last_msg_send_time), 0) AS msg_send_time," +
		"COALESCE(EXTRACT(EPOCH FROM last_msg_receipt_time), 0) AS msg_recv_time, " +
		"COALESCE(EXTRACT(EPOCH FROM latest_end_time), 0) AS reported_time, " +
		"COALESCE(NULL::numeric, 0) AS apply_error_count, COALESCE(NULL::numeric, 0) AS sync_error_count " +
		"FROM pg_stat_subscription WHERE relid ISNULL;"

	postgresStatSubscriptionQuery16 = "SELECT s1.subid, s1.subname, COALESCE(s1.pid, 0) AS pid, " +
		"COALESCE(NULL::text, 'unknown') AS worker_type, " +
		"COALESCE(received_lsn - '0/0', 0) AS received_lsn, " +
		"COALESCE(latest_end_lsn - '0/0', 0) AS reported_lsn, " +
		"COALESCE(EXTRACT(EPOCH FROM last_msg_send_time), 0) AS msg_send_time," +
		"COALESCE(EXTRACT(EPOCH FROM last_msg_receipt_time), 0) AS msg_recv_time, " +
		"COALESCE(EXTRACT(EPOCH FROM latest_end_time), 0) AS reported_time, " +
		"s2.apply_error_count, s2.sync_error_count " +
		"FROM pg_stat_subscription s1 JOIN pg_stat_subscription_stats s2 ON s1.subid = s2.subid " +
		"WHERE s1.relid ISNULL;"

	postgresStatSubscriptionQueryLatest = "SELECT s1.subid, s1.subname, COALESCE(s1.pid, 0) AS pid, " +
		"COALESCE(s1.worker_type, 'unknown') AS worker_type, " +
		"COALESCE(received_lsn - '0/0', 0) AS received_lsn, " +
		"COALESCE(latest_end_lsn - '0/0', 0) AS reported_lsn, " +
		"COALESCE(EXTRACT(EPOCH FROM last_msg_send_time), 0) AS msg_send_time," +
		"COALESCE(EXTRACT(EPOCH FROM last_msg_receipt_time), 0) AS msg_recv_time, " +
		"COALESCE(EXTRACT(EPOCH FROM latest_end_time), 0) AS reported_time, " +
		"s2.apply_error_count, s2.sync_error_count " +
		"FROM pg_stat_subscription s1 JOIN pg_stat_subscription_stats s2 ON s1.subid = s2.subid" +
		"WHERE s1.relid ISNULL;"
)

// postgresStatSubscriptionCollector defines metric descriptors and stats store.
type postgresStatSubscriptionCollector struct {
	labelNames    []string
	received_lsn  typedDesc
	reported_lsn  typedDesc
	msg_send_time typedDesc
	msg_recv_time typedDesc
	reported_time typedDesc
	errorCount    typedDesc
}

// NewPostgresStatSubscriptionCollector returns a new Collector exposing postgres pg_stat_subscription stats.
// For details see https://www.postgresql.org/docs/current/monitoring-stats.html#MONITORING-PG-STAT-SUBSCRIPTION
func NewPostgresStatSubscriptionCollector(constLabels labels, settings model.CollectorSettings) (Collector, error) {
	var labelNames = []string{"subid", "subname", "worker_type"}

	return &postgresStatSubscriptionCollector{
		labelNames: labelNames,
		received_lsn: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_subscription", "received_lsn", "Last write-ahead log location received.", 0},
			prometheus.GaugeValue,
			labelNames, constLabels,
			settings.Filters,
		),
		reported_lsn: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_subscription", "reported_lsn", "Last write-ahead log location reported to origin WAL sender.", 0},
			prometheus.GaugeValue,
			labelNames, constLabels,
			settings.Filters,
		),
		msg_send_time: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_subscription", "msg_send_time", "Send time of last message received from origin WAL sender.", 0},
			prometheus.GaugeValue,
			labelNames, constLabels,
			settings.Filters,
		),
		msg_recv_time: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_subscription", "msg_recv_time", "Receipt time of last message received from origin WAL sender.", 0},
			prometheus.GaugeValue,
			labelNames, constLabels,
			settings.Filters,
		),
		reported_time: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_subscription", "reported_time", "Time of last write-ahead log location reported to origin WAL sender.", 0},
			prometheus.GaugeValue,
			labelNames, constLabels,
			settings.Filters,
		),
		errorCount: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_subscription", "error_count", "Number of times an error occurred (applying changes OR initial table synchronization).", 0},
			prometheus.CounterValue,
			[]string{"subid", "subname", "worker_type", "type"}, constLabels,
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
				if value, ok := stat.values["received_lsn"]; ok {
					ch <- c.received_lsn.newConstMetric(value, stat.SubId, stat.SubName, stat.WorkerType)
				}
				if value, ok := stat.values["reported_lsn"]; ok {
					ch <- c.reported_lsn.newConstMetric(value, stat.SubId, stat.SubName, stat.WorkerType)
				}
				if value, ok := stat.values["msg_send_time"]; ok {
					ch <- c.msg_send_time.newConstMetric(value, stat.SubId, stat.SubName, stat.WorkerType)
				}
				if value, ok := stat.values["msg_recv_time"]; ok {
					ch <- c.msg_recv_time.newConstMetric(value, stat.SubId, stat.SubName, stat.WorkerType)
				}
				if value, ok := stat.values["reported_time"]; ok {
					ch <- c.reported_time.newConstMetric(value, stat.SubId, stat.SubName, stat.WorkerType)
				}
				if value, ok := stat.values["apply_error_count"]; ok {
					ch <- c.errorCount.newConstMetric(value, stat.SubId, stat.SubName, stat.WorkerType, "apply")
				}
				if value, ok := stat.values["sync_error_count"]; ok {
					ch <- c.errorCount.newConstMetric(value, stat.SubId, stat.SubName, stat.WorkerType, "sync")
				}
			}
		}
	}

	return nil
}

// postgresSubscriptionStat represents per-subscription stats based on pg_stat_subscription.
type postgresSubscriptionStat struct {
	SubId      string // a subid
	SubName    string // a subscription name
	Pid        string // a pid
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
			case "subid":
				stat.SubId = row[i].String
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
			case "pid":
				s.values["pid"] = v
			case "received_lsn":
				s.values["received_lsn"] = v
			case "reported_lsn":
				s.values["reported_lsn"] = v
			case "msg_send_time":
				s.values["msg_send_time"] = v
			case "msg_recv_time":
				s.values["msg_recv_time"] = v
			case "reported_time":
				s.values["reported_time"] = v
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
