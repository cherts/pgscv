// Package pgscv is a pgSCV main helper
package pgscv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cherts/pgscv/internal/http"
	"github.com/cherts/pgscv/internal/log"
	"github.com/cherts/pgscv/internal/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	net_http "net/http"
	"strings"
	"sync"
)

// Start is the application's starting point.
func Start(ctx context.Context, config *Config) error {
	log.Debug("start application")

	serviceRepo := service.NewRepository()

	serviceConfig := service.Config{
		NoTrackMode:        config.NoTrackMode,
		ConnDefaults:       config.Defaults,
		ConnsSettings:      config.ServicesConnsSettings,
		DatabasesRE:        config.DatabasesRE,
		DisabledCollectors: config.DisableCollectors,
		CollectorsSettings: config.CollectorsSettings,
		CollectTopTable:    config.CollectTopTable,
		CollectTopIndex:    config.CollectTopIndex,
		CollectTopQuery:    config.CollectTopQuery,
		SkipConnErrorMode:  config.SkipConnErrorMode,
	}

	if len(config.ServicesConnsSettings) == 0 {
		return errors.New("no services defined")
	}

	// fulfill service repo using passed services
	serviceRepo.AddServicesFromConfig(serviceConfig)

	// setup exporters for all services
	err := serviceRepo.SetupServices(serviceConfig)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup

	errCh := make(chan error)
	defer close(errCh)

	// Start HTTP metrics listener.
	wg.Add(1)
	go func() {
		if err := runMetricsListener(ctx, config, serviceRepo); err != nil {
			errCh <- err
		}
		wg.Done()
	}()

	// Waiting for errors or context cancelling.
	for {
		select {
		case <-ctx.Done():
			log.Info("exit signaled, stop application")
			cancel()
			wg.Wait()
			return nil
		case e := <-errCh:
			cancel()
			wg.Wait()
			return e
		}
	}
}

// getMetricsHandler return http handler function to /metrics endpoint
func getMetricsHandler(repository *service.Repository) func(w net_http.ResponseWriter, r *net_http.Request) {
	return func(w net_http.ResponseWriter, r *net_http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			h := promhttp.InstrumentMetricHandler(
				prometheus.DefaultRegisterer, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}),
			)
			h.ServeHTTP(w, r)
		} else {
			registry := repository.GetRegistry(target)
			if registry == nil {
				net_http.Error(w, fmt.Sprintf("Target %s not registered", target), http.StatusNotFound)
				return
			}
			h := promhttp.InstrumentMetricHandler(
				registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
			)
			h.ServeHTTP(w, r)
		}
	}
}

// getTargetsHandler return http handler function to /targets endpoint
func getTargetsHandler(repository *service.Repository, urlPrefix string, enableTLS bool) func(w net_http.ResponseWriter, r *net_http.Request) {
	return func(w net_http.ResponseWriter, r *net_http.Request) {
		svcIDs := repository.GetServiceIDs()
		targets := make([]string, len(svcIDs))
		var url string
		if urlPrefix != "" {
			url = strings.Trim(urlPrefix, "/")
		} else {
			if enableTLS {
				url = fmt.Sprintf("https://%s", r.Host)
			} else {
				url = r.Host
			}
		}
		for i, svcID := range svcIDs {
			targets[i] = fmt.Sprintf("%s/metrics?target=%s", url, svcID)
		}
		data := []struct {
			Targets []string `json:"targets"`
		}{
			0: {Targets: targets},
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			net_http.Error(w, err.Error(), net_http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonData)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

// runMetricsListener start HTTP listener accordingly to passed configuration.
func runMetricsListener(ctx context.Context, config *Config, repository *service.Repository) error {
	sCfg := http.ServerConfig{
		Addr:       config.ListenAddress,
		AuthConfig: config.AuthConfig,
	}
	srv := http.NewServer(sCfg, getMetricsHandler(repository), getTargetsHandler(repository, config.URLPrefix, config.AuthConfig.EnableTLS))

	errCh := make(chan error)
	defer close(errCh)

	// Run default listener.
	go func() {
		errCh <- srv.Serve()
	}()

	// Waiting for errors or context cancelling.
	for {
		select {
		case <-ctx.Done():
			log.Info("exit signaled, stop metrics listener")
			return nil
		case err := <-errCh:
			return err
		}
	}
}
