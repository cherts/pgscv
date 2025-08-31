package collector

import (
	"context"
	"strconv"
	"sync"

	"github.com/cherts/pgscv/internal/log"
	"github.com/cherts/pgscv/internal/model"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	postgresSubscriptionRel = `
		SELECT CURRENT_CATALOG AS datname, subname, srsubstate::TEXT AS state, count(*) AS count
		FROM pg_subscription_rel sr
		LEFT JOIN pg_stat_subscription ss ON sr.srsubid = ss.subid
		GROUP BY 2, 3;
`
)

// postgresSubscriptionRelCollector defines metric descriptors.
type postgresSubscriptionRelCollector struct {
	labelNames []string
	count      typedDesc
}

// NewPostgresSubscriptionRelCollector returns a new Collector exposing postgres pg_subscription_rel stats.
// For details see https://www.postgresql.org/docs/17/catalog-pg-subscription-rel.html#CATALOG-PG-SUBSCRIPTION-REL
func NewPostgresSubscriptionRelCollector(constLabels labels, settings model.CollectorSettings) (Collector, error) {
	var labelNames = []string{"datname", "subname", "state"}
	return &postgresSubscriptionRelCollector{
		labelNames: labelNames,
		count: newBuiltinTypedDesc(
			descOpts{"postgres", "subscription_rel", "count", "Count tables in replication state", 0},
			prometheus.GaugeValue,
			labelNames, constLabels,
			settings.Filters,
		),
	}, nil
}

// Update method collects statistics, parse it and produces metrics that are sent to Prometheus.
func (c *postgresSubscriptionRelCollector) Update(ctx context.Context, config Config, ch chan<- prometheus.Metric) error {
	conn := config.DB
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	var err error

	cacheKey, res := getFromCache(config.CacheConfig, config.ConnString, collectorPostgresSubscriptionRel, postgresSubscriptionRel)
	if res == nil {
		res, err = conn.Query(ctx, postgresSubscriptionRel)
		if err != nil {
			log.Warnf("get pg_subscription_rel failed: %s; skip", err)
			return err
		}
		saveToCache(collectorPostgresSubscriptionRel, wg, config.CacheConfig, cacheKey, res)
	}

	for _, row := range res.Rows {
		var (
			datName string
			subName string
			state   string
			count   float64
		)
		for i, colname := range res.Colnames {
			switch colname.Name {
			case "datname":
				datName = row[i].String
			case "subname":
				subName = row[i].String
			case "state":
				switch row[i].String {
				case "i":
					state = "initialize"
				case "d":
					state = "data_is_being_copied,"
				case "f":
					state = "finished_table_copy"
				case "s":
					state = "synchronized"
				case "r":
					state = "normal_replication"
				}
			case "count":
				count, err = strconv.ParseFloat(row[i].String, 64)
				if err != nil {
					return err
				}
			}
		}
		ch <- c.count.newConstMetric(count, datName, subName, state)
	}
	return nil
}
