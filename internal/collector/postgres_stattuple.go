package collector

import (
	"context"
	"fmt"
	"github.com/cherts/pgscv/internal/log"
	"github.com/cherts/pgscv/internal/model"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"sync"
)

const (
	postgresStatTupleQuery = `
		select current_database() as datname,
			   relname,
			   schemaname,
			   approx_tuple_percent,
			   dead_tuple_count,
			   dead_tuple_len,
			   dead_tuple_percent,
			   approx_free_space,
			   approx_free_percent
		from pg_stat_user_tables, %s.pgstattuple_approx(relid) stt
		where pg_relation_size(relid) > 50 * 1024 * 1024
`
)

type postgresStatTupleCollector struct {
	labelNames         []string
	approxTuplePercent typedDesc
	deadTupleCount     typedDesc
	deadTupleLen       typedDesc
	deadTuplePercent   typedDesc
	approxFreeSpace    typedDesc
	approxFreePercent  typedDesc
}

// NewPostgresStatTupleCollector returns a new Collector exposing postgres tuple-level statistics.
// For details see https://www.postgresql.org/docs/current/pgstattuple.html
func NewPostgresStatTupleCollector(constLabels labels, settings model.CollectorSettings) (Collector, error) {
	var labelNames = []string{"datname", "relname", "schemaname"}
	return &postgresStatTupleCollector{
		labelNames: labelNames,
		approxTuplePercent: newBuiltinTypedDesc(
			descOpts{"postgres", "pgstattuple", "approx_tuple_percent", "Percentage of live tuples", 0},
			prometheus.GaugeValue,
			labelNames, constLabels,
			settings.Filters,
		),
		deadTupleCount: newBuiltinTypedDesc(
			descOpts{"postgres", "pgstattuple", "dead_tuple_count", "Number of dead tuples (exact)", 0},
			prometheus.GaugeValue,
			labelNames, constLabels,
			settings.Filters,
		),
		deadTupleLen: newBuiltinTypedDesc(
			descOpts{"postgres", "pgstattuple", "dead_tuple_len", "Total length of dead tuples in bytes (exact)", 0},
			prometheus.GaugeValue,
			labelNames, constLabels,
			settings.Filters,
		),
		deadTuplePercent: newBuiltinTypedDesc(
			descOpts{"postgres", "pgstattuple", "dead_tuple_percent", "Percentage of dead tuples", 0},
			prometheus.GaugeValue,
			labelNames, constLabels,
			settings.Filters,
		),
		approxFreeSpace: newBuiltinTypedDesc(
			descOpts{"postgres", "pgstattuple", "approx_free_space", "Total free space in bytes (estimated)", 0},
			prometheus.GaugeValue,
			labelNames, constLabels,
			settings.Filters,
		),
		approxFreePercent: newBuiltinTypedDesc(
			descOpts{"postgres", "pgstattuple", "approx_free_percent", "Percentage of free space", 0},
			prometheus.GaugeValue,
			labelNames, constLabels,
			settings.Filters,
		),
	}, nil
}

// Update method collects statistics, parse it and produces metrics that are sent to Prometheus.
func (c *postgresStatTupleCollector) Update(ctx context.Context, config Config, ch chan<- prometheus.Metric) error {
	if !config.pgStatTuple {
		return nil
	}

	conn := config.DB
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	var err error
	query := selectStatTupleQuery(config.pgStatTupleSchema)
	cacheKey, res := getFromCache(config.CacheConfig, config.ConnString, collectorPostgresStatTuple, query)
	if res == nil {
		res, err = conn.Query(ctx, query)
		if err != nil {
			log.Warnf("get pg_subscription_rel failed: %s; skip", err)
			return err
		}
		saveToCache(collectorPostgresStatTuple, wg, config.CacheConfig, cacheKey, res)
	}

	for _, row := range res.Rows {
		var (
			datName            string
			relName            string
			schemaName         string
			approxTuplePercent float64
			deadTupleCount     float64
			deadTupleLen       float64
			deadTuplePercent   float64
			approxFreeSpace    float64
			approxFreePercent  float64
		)
		for i, colname := range res.Colnames {
			switch colname.Name {
			case "datname":
				datName = row[i].String
			case "relname":
				relName = row[i].String
			case "schemaname":
				schemaName = row[i].String
			case "approx_tuple_percent":
				approxTuplePercent, err = strconv.ParseFloat(row[i].String, 64)
				if err != nil {
					return err
				}
			case "dead_tuple_count":
				deadTupleCount, err = strconv.ParseFloat(row[i].String, 64)
				if err != nil {
					return err
				}
			case "dead_tuple_len":
				deadTupleLen, err = strconv.ParseFloat(row[i].String, 64)
				if err != nil {
					return err
				}
			case "dead_tuple_percent":
				deadTuplePercent, err = strconv.ParseFloat(row[i].String, 64)
				if err != nil {
					return err
				}
			case "approx_free_space":
				approxFreeSpace, err = strconv.ParseFloat(row[i].String, 64)
				if err != nil {
					return err
				}
			case "approx_free_percent":
				approxFreePercent, err = strconv.ParseFloat(row[i].String, 64)
				if err != nil {
					return err
				}
			}
		}
		ch <- c.approxTuplePercent.newConstMetric(approxTuplePercent, datName, relName, schemaName)
		ch <- c.deadTupleCount.newConstMetric(deadTupleCount, datName, relName, schemaName)
		ch <- c.deadTupleLen.newConstMetric(deadTupleLen, datName, relName, schemaName)
		ch <- c.deadTuplePercent.newConstMetric(deadTuplePercent, datName, relName, schemaName)
		ch <- c.approxFreeSpace.newConstMetric(approxFreeSpace, datName, relName, schemaName)
		ch <- c.approxFreePercent.newConstMetric(approxFreePercent, datName, relName, schemaName)
	}
	return nil
}

// selectStatTupleQuery returns query with specified extension schema.
func selectStatTupleQuery(schema string) string {
	return fmt.Sprintf(postgresStatTupleQuery, schema)
}
