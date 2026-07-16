package collector

import (
	"fmt"
	"strconv"

	"github.com/cherts/pgscv/internal/log"
	"github.com/cherts/pgscv/internal/model"
	"github.com/cherts/pgscv/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	postgresSubscriptionRel15 = "SELECT CURRENT_CATALOG AS datname, subname, srsubstate::TEXT AS state, count(*) AS count " +
		"FROM pg_subscription_rel sr " +
		"LEFT JOIN pg_stat_subscription ss ON sr.srsubid = ss.subid " +
		"WHERE CURRENT_CATALOG = '%s' " +
		"GROUP BY 2, 3"

	postgresSubscriptionRelLatest = "SELECT CURRENT_CATALOG AS datname, subname, srsubstate::TEXT AS state, count(*) AS count " +
		"FROM pg_subscription_rel sr " +
		"LEFT JOIN pg_stat_subscription ss ON sr.srsubid = ss.subid " +
		"WHERE leader_pid IS NULL AND CURRENT_CATALOG = '%s' " +
		"GROUP BY 2, 3"
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
func (c *postgresSubscriptionRelCollector) Update(config Config, ch chan<- prometheus.Metric) error {
	if config.pgVersion.Numeric < PostgresV10 {
		log.Debugln("[postgres subscription_rel collector]: pg_subscription_rel view are not available, required Postgres 10 or newer")
		return nil
	}

	conn, err := store.New(config.ConnString, config.ConnTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	collect := func(conn *store.DB) {
		collectSubscriptionRel(conn, config, ch, c.count)
	}

	if config.DatabasesRE == nil {
		// service discovery case
		collect(conn)
		return nil
	}

	databases, err := listDatabases(conn)
	if err != nil {
		return err
	}

	pgconfig, err := pgx.ParseConfig(config.ConnString)
	if err != nil {
		return err
	}

	// walk through all databases, connect to it and collect schema-specific stats
	for _, d := range databases {
		// Skip database if not matched to allowed.
		if !config.DatabasesRE.MatchString(d) {
			continue
		}
		pgconfig.Database = d
		conn, err := store.NewWithConfig(pgconfig)
		if err != nil {
			return err
		}
		defer conn.Close()
		collect(conn)
	}

	return nil
}

// collectSubscriptionRel collects metrics related to pg_subscription_rel.
func collectSubscriptionRel(conn *store.DB, config Config, ch chan<- prometheus.Metric, desc typedDesc) {
	database := conn.Conn().Config().Database
	res, err := conn.Query(selectSubscriptionRelQuery(config.pgVersion.Numeric, database))
	if err != nil {
		log.Warnf("get pg_subscription_rel failed: %s; skip", err)
	} else {
		log.Debug("parse postgres subscription_rel stats")

		for _, row := range res.Rows {
			var (
				datName string
				subName string
				state   string
				count   float64
			)
			for i, colname := range res.Colnames {
				switch string(colname.Name) {
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
						continue
					}
				}
			}
			ch <- desc.newConstMetric(count, datName, subName, state)
		}

	}
}

// selectSubscriptionRelQuery returns suitable subscription_rel query depending on passed version.
func selectSubscriptionRelQuery(version int, database string) string {
	switch {
	case version < PostgresV16:
		return fmt.Sprintf(postgresSubscriptionRel15, database)
	default:
		return fmt.Sprintf(postgresSubscriptionRelLatest, database)
	}
}
