// Package pgscv is a pgSCV main helper
package pgscv

import (
	"context"
	"errors"
	sd "github.com/cherts/pgscv/internal/discovery/service"
	"github.com/cherts/pgscv/internal/model"
	"sync"

	"github.com/cherts/pgscv/internal/http"
	"github.com/cherts/pgscv/internal/log"
	"github.com/cherts/pgscv/internal/service"
)

const pgSCVSubscriber = "pgscv_subscriber"

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
		ConnTimeout:        config.ConnTimeout,
	}

	if len(config.ServicesConnsSettings) == 0 && config.DiscoveryServices == nil {
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

	errCh := make(chan error, 2)
	defer close(errCh)
	if config.DiscoveryServices != nil {
		for _, ds := range *config.DiscoveryServices {
			wg.Add(1)
			go func() {
				err := ds.Start(ctx, errCh)
				if err != nil {
					errCh <- err
				}
				wg.Done()
			}()
			switch dt := ds.(type) {
			case *sd.YandexDiscovery:
				err := subscribeYandex(&ds, config, serviceRepo)
				if err != nil {
					cancel()
					return err
				}
			default:
				log.Infof("unknown discovery type %T", dt)
			}

		}
	}

	// Start HTTP metrics listener.
	wg.Add(1)
	go func() {
		if err := runMetricsListener(ctx, config); err != nil {
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

func subscribeYandex(ds *sd.Discovery, config *Config, serviceRepo *service.Repository) error {
	err := (*ds).Subscribe(pgSCVSubscriber,
		// addService
		func(services map[string]sd.Service) error {
			constLabels := make(map[string]*map[string]string)
			serviceDiscoveryConfig := service.Config{
				NoTrackMode:        config.NoTrackMode,
				ConnDefaults:       config.Defaults,
				DisabledCollectors: config.DisableCollectors,
				CollectorsSettings: config.CollectorsSettings,
				CollectTopTable:    config.CollectTopTable,
				CollectTopIndex:    config.CollectTopIndex,
				CollectTopQuery:    config.CollectTopQuery,
				SkipConnErrorMode:  config.SkipConnErrorMode,
				ConstLabels:        &constLabels,
			}
			var cs = make(service.ConnsSettings, len(services))
			for serviceID, svc := range services {
				cs[serviceID] = service.ConnSetting{
					ServiceType: model.ServiceTypePostgresql,
					Conninfo:    svc.DSN,
				}
				constLabels[serviceID] = &svc.ConstLabels
			}
			serviceDiscoveryConfig.ConnsSettings = cs
			serviceRepo.AddServicesFromConfig(serviceDiscoveryConfig)
			err := serviceRepo.SetupServices(serviceDiscoveryConfig)
			if err != nil {
				return err
			}
			return nil
		},
		// removeService
		func(serviceIds []string) error {
			for _, serviceID := range serviceIds {
				log.Infof("unregister service [%s]", serviceID)
				serviceRepo.RemoveService(serviceID)
			}
			return nil
		},
	)
	return err
}

// runMetricsListener start HTTP listener accordingly to passed configuration.
func runMetricsListener(ctx context.Context, config *Config) error {
	srv := http.NewServer(http.ServerConfig{
		Addr:       config.ListenAddress,
		AuthConfig: config.AuthConfig,
	})

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
