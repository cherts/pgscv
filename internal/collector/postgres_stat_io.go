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
	postgresStatIoQuery = "SELECT backend_type, object, context, coalesce(reads, 0), coalesce(read_time, 0), coalesce(writes, 0), " +
		"coalesce(write_time, 0), coalesce(writebacks, 0), coalesce(writeback_time, 0), coalesce(extends, 0), coalesce(extend_time, 0), " +
		"coalesce(op_bytes, 0), coalesce(hits, 0), coalesce(evictions, 0), coalesce(reuses, 0), coalesce(fsyncs, 0), coalesce(fsync_time, 0) " +
		"FROM pg_stat_io"
)

// postgresStatIOCollector defines metric descriptors and stats store.
type postgresStatIOCollector struct {
	reads          typedDesc
	read_time      typedDesc
	writes         typedDesc
	write_time     typedDesc
	writebacks     typedDesc
	writeback_time typedDesc
	extends        typedDesc
	extend_time    typedDesc
	hits           typedDesc
	evictions      typedDesc
	reuses         typedDesc
	fsyncs         typedDesc
	fsync_time     typedDesc
	labelNames     []string
}

// NewPostgresStatIOCollector returns a new Collector exposing postgres pg_stat_io stats.
func NewPostgresStatIOCollector(constLabels labels, settings model.CollectorSettings) (Collector, error) {
	var labels = []string{"backend_type", "object", "context"}

	return &postgresStatIOCollector{
		reads: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_io", "reads", "Labeled info about reads.", 0},
			prometheus.GaugeValue,
			labels, constLabels,
			settings.Filters,
		),
		read_time: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_io", "read_time", "Labeled info about read_time.", 0},
			prometheus.GaugeValue,
			labels, constLabels,
			settings.Filters,
		),
		writes: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_io", "writes", "Labeled info about writes.", 0},
			prometheus.GaugeValue,
			labels, constLabels,
			settings.Filters,
		),
		write_time: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_io", "write_time", "Labeled info about write_time.", 0},
			prometheus.GaugeValue,
			labels, constLabels,
			settings.Filters,
		),
		writebacks: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_io", "writebacks", "Labeled info about writebacks.", 0},
			prometheus.GaugeValue,
			labels, constLabels,
			settings.Filters,
		),
		writeback_time: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_io", "writeback_time", "Labeled info about writeback_time.", 0},
			prometheus.GaugeValue,
			labels, constLabels,
			settings.Filters,
		),
		extends: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_io", "extends", "Labeled info about extends.", 0},
			prometheus.GaugeValue,
			labels, constLabels,
			settings.Filters,
		),
		extend_time: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_io", "extend_time", "Labeled info about extend_time.", 0},
			prometheus.GaugeValue,
			labels, constLabels,
			settings.Filters,
		),
		hits: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_io", "hits", "Labeled info about hits.", 0},
			prometheus.GaugeValue,
			labels, constLabels,
			settings.Filters,
		),
		evictions: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_io", "evictions", "Labeled info about evictions.", 0},
			prometheus.GaugeValue,
			labels, constLabels,
			settings.Filters,
		),
		reuses: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_io", "reuses", "Labeled info about reuses.", 0},
			prometheus.GaugeValue,
			labels, constLabels,
			settings.Filters,
		),
		fsyncs: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_io", "fsyncs", "Labeled info about fsyncs.", 0},
			prometheus.GaugeValue,
			labels, constLabels,
			settings.Filters,
		),
		fsync_time: newBuiltinTypedDesc(
			descOpts{"postgres", "stat_io", "fsync_time", "Labeled info about fsync_time.", 0},
			prometheus.GaugeValue,
			labels, constLabels,
			settings.Filters,
		),
	}, nil
}

// Update method collects statistics, parse it and produces metrics that are sent to Prometheus.
func (c *postgresStatIOCollector) Update(config Config, ch chan<- prometheus.Metric) error {
	if config.serverVersionNum < PostgresV16 {
		log.Debugln("[postgres stat_io collector]: some server-side functions are not available, required Postgres 16 or newer")
		return nil
	}

	conn, err := store.New(config.ConnString, config.ConnTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Collecting pg_stst_io since Postgres 16.
	if config.serverVersionNum >= PostgresV16 {
		res, err := conn.Query(postgresStatIoQuery)
		if err != nil {
			log.Warnf("get pg_stat_io failed: %s; skip", err)
		} else {
			stats := parsePostgresStatIO(res, []string{"backend_type", "object", "context"})

			for _, stat := range stats {
				ch <- c.reads.newConstMetric(stat.Reads, stat.BackendType, stat.IoContext, stat.IoObject)
				ch <- c.read_time.newConstMetric(stat.ReadTime, stat.BackendType, stat.IoContext, stat.IoObject)
				ch <- c.writes.newConstMetric(stat.Writes, stat.BackendType, stat.IoContext, stat.IoObject)
				ch <- c.writebacks.newConstMetric(stat.Writebacks, stat.BackendType, stat.IoContext, stat.IoObject)
				ch <- c.writeback_time.newConstMetric(stat.WritebackTime, stat.BackendType, stat.IoContext, stat.IoObject)
				ch <- c.extends.newConstMetric(stat.Extends, stat.BackendType, stat.IoContext, stat.IoObject)
				ch <- c.extend_time.newConstMetric(stat.ExtendTime, stat.BackendType, stat.IoContext, stat.IoObject)
				ch <- c.hits.newConstMetric(stat.Hits, stat.BackendType, stat.IoContext, stat.IoObject)
				ch <- c.evictions.newConstMetric(stat.Evictions, stat.BackendType, stat.IoContext, stat.IoObject)
				ch <- c.reuses.newConstMetric(stat.Reuses, stat.BackendType, stat.IoContext, stat.IoObject)
				ch <- c.fsyncs.newConstMetric(stat.Fsyncs, stat.BackendType, stat.IoContext, stat.IoObject)
				ch <- c.fsync_time.newConstMetric(stat.FsyncTime, stat.BackendType, stat.IoContext, stat.IoObject)
			}
		}
	}

	return nil
}

// postgresStatIO
type postgresStatIO struct {
	BackendType   string // a backend type like "autovacuum worker"
	IoObject      string // "relation" or "temp relation"
	IoContext     string // "normal", "vacuum", "bulkread" or "bulkwrite"
	Reads         float64
	ReadTime      float64
	Writes        float64
	WriteTime     float64
	Writebacks    float64
	WritebackTime float64
	Extends       float64
	ExtendTime    float64
	OpBytes       float64
	Hits          float64
	Evictions     float64
	Reuses        float64
	Fsyncs        float64
	FsyncTime     float64
}

// parsePostgresStatIO parses PGResult and returns structs with stats values.
func parsePostgresStatIO(r *model.PGResult, labelNames []string) map[string]postgresStatIO {
	log.Debug("parse postgres stat_io stats")

	var stat_io = make(map[string]postgresStatIO)

	for _, row := range r.Rows {
		var BackendType, IoObject, IoContext string

		for i, colname := range r.Colnames {
			switch string(colname.Name) {
			case "backend_type":
				BackendType = row[i].String
			case "object":
				IoObject = row[i].String
			case "context":
				IoContext = row[i].String
			}
		}

		// create a stat_io name consisting of trio BackendType/IoObject/IoContext
		stats := strings.Join([]string{BackendType, IoObject, IoContext}, "/")

		// Put stats with labels (but with no data values yet) into stats store.
		if _, ok := stat_io[stats]; !ok {
			stat_io[stats] = postgresStatIO{BackendType: BackendType, IoObject: IoObject, IoContext: IoContext}
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

			s := stat_io[stats]

			switch string(colname.Name) {
			case "reads":
				s.Reads = v
			case "read_time":
				s.ReadTime = v
			case "writes":
				s.Writes = v
			case "write_time":
				s.WriteTime = v
			case "writebacks":
				s.Writebacks = v
			case "writeback_time":
				s.WritebackTime = v
			case "extends":
				s.Extends = v
			case "extend_time":
				s.ExtendTime = v
			case "op_bytes":
				s.OpBytes = v
			case "hits":
				s.Hits = v
			case "evictions":
				s.Evictions = v
			case "fsyncs":
				s.Fsyncs = v
			case "fsync_time":
				s.FsyncTime = v
			default:
				continue
			}

			stat_io[stats] = s
		}
	}

	return stat_io
}
