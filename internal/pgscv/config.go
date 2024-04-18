package pgscv

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cherts/pgscv/internal/http"
	"github.com/cherts/pgscv/internal/log"
	"github.com/cherts/pgscv/internal/model"
	"github.com/cherts/pgscv/internal/service"
	"github.com/jackc/pgx/v4"
	"gopkg.in/yaml.v2"
)

const (
	defaultListenAddress     = "127.0.0.1:9890"
	defaultPostgresUsername  = "pgscv"
	defaultPostgresDbname    = "postgres"
	defaultPgbouncerUsername = "pgscv"
	defaultPgbouncerDbname   = "pgbouncer"
)

// Config defines application's configuration.
type Config struct {
	NoTrackMode           bool                     `yaml:"no_track_mode"`      // controls tracking sensitive information (query texts, etc)
	ListenAddress         string                   `yaml:"listen_address"`     // Network address and port where the application should listen on
	ServicesConnsSettings service.ConnsSettings    `yaml:"services"`           // All connections settings for exact services
	Defaults              map[string]string        `yaml:"defaults"`           // Defaults
	DisableCollectors     []string                 `yaml:"disable_collectors"` // List of collectors which should be disabled. DEPRECATED in favor collectors settings
	CollectorsSettings    model.CollectorsSettings `yaml:"collectors"`         // Collectors settings propagated from main YAML configuration
	Databases             string                   `yaml:"databases"`          // Regular expression string specifies databases from which metrics should be collected
	DatabasesRE           *regexp.Regexp           // Regular expression object compiled from Databases
	AuthConfig            http.AuthConfig          `yaml:"authentication"` // TLS and Basic auth configuration
}

// NewConfig creates new config based on config file or return default config if config file is not specified.
func NewConfig(configFilePath string) (*Config, error) {
	// Получить конфигурацию из файла
	var configFromFile *Config
	if configFilePath != "" {
		configRealPath, err := RealPath(configFilePath)
		if err != nil {
			return nil, err
		}
		log.Infoln("read configuration from ", configRealPath)
		content, err := os.ReadFile(filepath.Clean(configRealPath))
		if err != nil {
			return nil, err
		}
		configFromFile = &Config{Defaults: map[string]string{}}
		err = yaml.Unmarshal(content, configFromFile)
		if err != nil {
			return nil, err
		}
	}

	// Получить конфигурацию из переменных окружения
	configFromEnv, err := newConfigFromEnv()
	if err != nil {
		return nil, err
	}

   // Объединить значения из configFromFile и configFromEnv
   if configFromFile != nil {
	// Обновляем значения из configFromEnv в configFromFile
	configFromFile.NoTrackMode = configFromEnv.NoTrackMode
	if configFromEnv.ListenAddress != "" {
		configFromFile.ListenAddress = configFromEnv.ListenAddress
	}
	if !reflect.DeepEqual(configFromEnv.ServicesConnsSettings, service.ConnsSettings{}) {
		configFromFile.ServicesConnsSettings = configFromEnv.ServicesConnsSettings
	}
	for key, value := range configFromEnv.Defaults {
		configFromFile.Defaults[key] = value
	}
	configFromFile.DisableCollectors = append(configFromFile.DisableCollectors, configFromEnv.DisableCollectors...)
	if !reflect.DeepEqual(configFromEnv.CollectorsSettings, model.CollectorsSettings{}) {
		configFromFile.CollectorsSettings = configFromEnv.CollectorsSettings
	}
	// Устанавливаем нужные значения в поле Databases
	if configFromFile.Databases == "" {
		configFromFile.Databases = configFromEnv.Databases
	}
	// Устанавливаем нужные значения в поле DatabasesRE
	if configFromEnv.DatabasesRE != nil {
		configFromFile.DatabasesRE = configFromEnv.DatabasesRE
	}
	// Устанавливаем нужные значения в поле AuthConfig
	if configFromEnv.AuthConfig != (http.AuthConfig{}) {
		configFromFile.AuthConfig = configFromEnv.AuthConfig
	}
	return configFromFile, nil
}

return configFromEnv, nil
}

// Read real config file path
func RealPath(filePath string) (string, error) {
	log.Infoln("reading file information ", filePath)
	fileInfo, err := os.Lstat(filepath.Clean(filePath))
	if err != nil {
		return filePath, err
	}
	if fileInfo.Mode()&fs.ModeSymlink != 0 {
		log.Debugln("is symlink")
		link, err := filepath.EvalSymlinks(filePath)
		if err != nil {
			return filePath, err
		}
		log.Debugln("resolved symlink to ", link)
		return link, nil
	} else if fileInfo.Mode().IsRegular() {
		log.Debugln("is regular file")
		return filePath, nil
	} else if fileInfo.Mode()&fs.ModeNamedPipe != 0 {
		log.Debugln("is named pipe")
		return filePath, nil
	} else if fileInfo.Mode().IsDir() {
		log.Debugln("is directory")
		return filePath, err
	} else {
		return filePath, err
	}
}

// Validate checks configuration for stupid values and set defaults
func (c *Config) Validate() error {
	if c.ListenAddress == "" {
		c.ListenAddress = defaultListenAddress
	}

	if c.NoTrackMode {
		log.Infoln("no-track enabled for [pg_stat_statements.query].")
	} else {
		log.Infoln("no-track disabled, for details check the documentation about 'no_track_mode' option.")
	}

	// setup defaults
	if c.Defaults == nil {
		c.Defaults = map[string]string{}
	}

	if _, ok := c.Defaults["postgres_username"]; !ok {
		c.Defaults["postgres_username"] = defaultPostgresUsername
	}

	if _, ok := c.Defaults["postgres_dbname"]; !ok {
		c.Defaults["postgres_dbname"] = defaultPostgresDbname
	}

	if _, ok := c.Defaults["pgbouncer_username"]; !ok {
		c.Defaults["pgbouncer_username"] = defaultPgbouncerUsername
	}

	if _, ok := c.Defaults["pgbouncer_dbname"]; !ok {
		c.Defaults["pgbouncer_dbname"] = defaultPgbouncerDbname
	}

	// User might specify its own set of services which he would like to monitor. This services should be validated and
	// invalid should be rejected. Validation is performed using pgx.ParseConfig method which does all dirty work.
	if c.ServicesConnsSettings != nil {
		if len(c.ServicesConnsSettings) != 0 {
			for k, s := range c.ServicesConnsSettings {
				if k == "" {
					return fmt.Errorf("empty service specified")
				}
				if s.ServiceType == "" {
					return fmt.Errorf("empty service_type for %s", k)
				}

				_, err := pgx.ParseConfig(s.Conninfo)
				if err != nil {
					return fmt.Errorf("invalid conninfo for %s: %s", k, err)
				}
			}
		}
	}

	// Create 'databases' regexp object for builtin metrics.
	re, err := newDatabasesRegexp(c.Databases)
	if err != nil {
		return err
	}
	c.DatabasesRE = re

	// Validate collector settings.
	err = validateCollectorSettings(c.CollectorsSettings)
	if err != nil {
		return err
	}

	// Validate authentication settings.
	enableAuth, enableTLS, err := c.AuthConfig.Validate()
	if err != nil {
		return err
	}
	c.AuthConfig.EnableAuth = enableAuth
	c.AuthConfig.EnableTLS = enableTLS

	return nil
}

// validateCollectorSettings validates collectors settings passed from main YAML configuration.
func validateCollectorSettings(cs model.CollectorsSettings) error {
	if len(cs) == 0 {
		return nil
	}

	for csName, settings := range cs {
		re1 := regexp.MustCompile(`^[a-zA-Z0-9]+/[a-zA-Z0-9]+$`)
		if !re1.MatchString(csName) {
			return fmt.Errorf("invalid collector name: %s", csName)
		}

		err := settings.Filters.Compile()
		if err != nil {
			return err
		}

		// Validate subsystems level
		for ssName, subsys := range settings.Subsystems {
			re2 := regexp.MustCompilePOSIX(`^[a-zA-Z0-9_]+$`)

			if !re2.MatchString(ssName) {
				return fmt.Errorf("invalid subsystem name: %s", ssName)
			}

			// Validate databases regexp.
			_, err := regexp.Compile(subsys.Databases)
			if err != nil {
				return fmt.Errorf("databases invalid regular expression specified: %s", err)
			}

			// Query must be specified if any metrics.
			if len(subsys.Metrics) > 0 && subsys.Query == "" {
				return fmt.Errorf("query is not specified for subsystem '%s' metrics", ssName)
			}

			// Validate metrics level
			reMetric := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

			for _, m := range subsys.Metrics {
				if m.Value == "" && m.LabeledValues == nil {
					return fmt.Errorf("value or labeled_values should be specified for metric '%s'", m.ShortName)
				}

				if m.Value != "" && m.LabeledValues != nil {
					return fmt.Errorf("value and labeled_values cannot be used together for metric '%s'", m.ShortName)
				}

				usage := m.Usage
				switch usage {
				case "COUNTER", "GAUGE":
					if !reMetric.MatchString(m.ShortName) {
						return fmt.Errorf("invalid metric name '%s'", m.ShortName)
					}
					if m.Description == "" {
						return fmt.Errorf("metric description is not specified for %s", m.ShortName)
					}
				default:
					return fmt.Errorf("invalid metric usage '%s'", usage)
				}
			}
		}
	}

	return nil
}

// newConfigFromEnv create config using environment variables.
func newConfigFromEnv() (*Config, error) {
	log.Infoln("read configuration from environment")

	config := &Config{
		Defaults:              map[string]string{},
		ServicesConnsSettings: map[string]service.ConnSetting{},
	}

	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "PGSCV_") &&
			!strings.HasPrefix(env, "POSTGRES_DSN") &&
			!strings.HasPrefix(env, "DATABASE_DSN") &&
			!strings.HasPrefix(env, "PGBOUNCER_DSN") &&
			!strings.HasPrefix(env, "PATRONI_URL") {
			continue
		}

		ff := strings.SplitN(env, "=", 2)

		key, value := ff[0], ff[1]

		// Parse POSTGRES_DSN (or its alias DATABASE_DSN).
		if strings.HasPrefix(key, "POSTGRES_DSN") || strings.HasPrefix(key, "DATABASE_DSN") {
			id, cs, err := service.ParsePostgresDSNEnv(key, value)
			if err != nil {
				return nil, err
			}

			config.ServicesConnsSettings[id] = cs
		}

		// Parse PGBOUNCER_DSN.
		if strings.HasPrefix(key, "PGBOUNCER_DSN") {
			id, cs, err := service.ParsePgbouncerDSNEnv(key, value)
			if err != nil {
				return nil, err
			}

			config.ServicesConnsSettings[id] = cs
		}

		// Parse PATRONI_URL.
		if strings.HasPrefix(key, "PATRONI_URL") {
			id, cs, err := service.ParsePatroniURLEnv(key, value)
			if err != nil {
				return nil, err
			}

			config.ServicesConnsSettings[id] = cs
		}

		switch key {
		case "PGSCV_LISTEN_ADDRESS":
			config.ListenAddress = value
		case "PGSCV_NO_TRACK_MODE":
			switch value {
			case "y", "yes", "Yes", "YES", "t", "true", "True", "TRUE", "1", "on":
				config.NoTrackMode = true
			default:
				config.NoTrackMode = false
			}
		case "PGSCV_DATABASES":
			config.Databases = value
		case "PGSCV_DISABLE_COLLECTORS":
			config.DisableCollectors = strings.Split(strings.Replace(value, " ", "", -1), ",")
		case "PGSCV_AUTH_USERNAME":
			config.AuthConfig.Username = value
		case "PGSCV_AUTH_PASSWORD":
			config.AuthConfig.Password = value
		case "PGSCV_AUTH_KEYFILE":
			config.AuthConfig.Keyfile = value
		case "PGSCV_AUTH_CERTFILE":
			config.AuthConfig.Certfile = value
		}
	}

	return config, nil
}

// newDatabasesRegexp creates new regexp depending on passed string.
func newDatabasesRegexp(s string) (*regexp.Regexp, error) {
	if s == "" {
		s = ".+"
	}

	return regexp.Compile(s)
}
