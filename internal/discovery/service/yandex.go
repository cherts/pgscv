package service

import (
	"context"
	"github.com/cherts/pgscv/internal/discovery/cloud/yandex"
	"github.com/cherts/pgscv/internal/discovery/mapops"
	"github.com/cherts/pgscv/internal/log"
	"gopkg.in/yaml.v2"
	"maps"
	"sync"
	"time"
)

type cluster struct {
	Name string `yaml:"name" json:"name,required"`
	// when Db is nil, iterate over all db's
	Db          *string `yaml:"db" json:"db"`
	ExcludeName *string `yaml:"exclude_name" json:"exclude_name"`
	ExcludeDb   *string `yaml:"exclude_db" json:"exclude_db"`
}

// YandexConfig AuthorizedKey - path to json file, Password - password for
// databases, Clusters - array of structures with matching rules
type YandexConfig struct {
	AuthorizedKey   string    `json:"authorized_key" yaml:"authorized_key"`
	FolderID        string    `json:"folder_id" yaml:"folder_id"`
	User            string    `json:"user" yaml:"user"`
	Password        string    `json:"password" yaml:"password"`
	RefreshInterval int       `json:"refresh_interval" yaml:"refresh_interval"`
	Clusters        []cluster `json:"clusters" yaml:"clusters"`
}

type engineIdx int
type version uint64

type subscriber struct {
	AddService     AddServiceFunc
	RemoveService  RemoveServiceFunc
	SyncedVersion  map[engineIdx]version
	syncedServices map[string]Service
}

// YandexDiscovery is main struct for Yandex Managed Databases discoverer
type YandexDiscovery struct {
	sync.RWMutex
	config      []YandexConfig
	engines     []*yandexEngine
	subscribers map[string]subscriber
}

// Unsubscribe implementation Unsubscribe method of Discovery interface
func (yd *YandexDiscovery) Unsubscribe(subscriberID string) error {
	yd.Lock()
	defer yd.Unlock()
	if _, ok := yd.subscribers[subscriberID]; !ok {
		return nil
	}
	svc := make([]string, 0, len(yd.subscribers[subscriberID].syncedServices))
	for k := range maps.Keys(yd.subscribers[subscriberID].syncedServices) {
		svc = append(svc, k)
	}
	err := yd.subscribers[subscriberID].RemoveService(svc)
	delete(yd.subscribers, subscriberID)
	return err
}

// Subscribe implementation Subscribe method of Discovery interface
func (yd *YandexDiscovery) Subscribe(subscriberID string, addService AddServiceFunc, removeService RemoveServiceFunc) error {
	yd.Lock()
	defer yd.Unlock()
	log.Debugf("YCD Subscribe %s", subscriberID)
	yd.subscribers[subscriberID] = subscriber{AddService: addService, RemoveService: removeService, syncedServices: make(map[string]Service), SyncedVersion: make(map[engineIdx]version)}
	for engineID, e := range yd.engines {
		e.RLock()
		for serviceID, svc := range e.dsn {
			labels := make(map[string]string)
			labels["mdb_cluster"] = svc.name
			yd.subscribers[subscriberID].syncedServices[string(serviceID)] = Service{DSN: svc.dsn, ConstLabels: labels}
		}
		e.RUnlock()
		yd.subscribers[subscriberID].SyncedVersion[engineIdx(engineID)] = e.version
	}
	if len(yd.subscribers[subscriberID].syncedServices) > 0 {
		err := addService(yd.subscribers[subscriberID].syncedServices)
		if err != nil {
			log.Errorf("Error adding synced services: %v", err)
		}
		return err
	}
	return nil
}

// Sync implementation Sync method of Discovery interface
func (yd *YandexDiscovery) Sync() error {
	yd.Lock()
	defer yd.Unlock()
	log.Debug("YCD Sync")
	for _, subscriber := range yd.subscribers {
		needSync := false
		engineServices := make(map[string]clusterDSN)
		for engineID, e := range yd.engines {
			e.RLock()
			if !(subscriber.SyncedVersion[engineIdx(engineID)] == e.version) {
				subscriber.SyncedVersion[engineIdx(engineID)] = e.version
				needSync = true
			}
			for serviceID, dsn := range e.dsn {
				engineServices[string(serviceID)] = dsn
			}
			e.RUnlock()
		}
		if !needSync {
			continue
		}

		removeSvc := make([]string, 0)
		appendSvc := make(map[string]Service)
		for _, v := range mapops.FullJoin(engineServices, subscriber.syncedServices) {
			if v.Left == nil {
				removeSvc = append(removeSvc, *v.Right)
				delete(subscriber.syncedServices, *v.Right)
			}
			if v.Right == nil {
				labels := make(map[string]string)
				labels["mdb_cluster"] = engineServices[*v.Left].name
				labels["provider"] = yandexMDB
				appendSvc[*v.Left] = Service{DSN: engineServices[*v.Left].dsn, ConstLabels: labels}
				subscriber.syncedServices[*v.Left] = appendSvc[*v.Left]
			}
		}
		if len(removeSvc) > 0 {
			log.Debugf("YCD removing %d services from subscriber", len(removeSvc))
			err := subscriber.RemoveService(removeSvc)
			if err != nil {
				return err
			}
		}
		if len(appendSvc) > 0 {
			log.Debugf("YCD appending %d services to sunscriber", len(appendSvc))
			err := subscriber.AddService(appendSvc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Init implementation Init method of Discovery interface
func (yd *YandexDiscovery) Init(cfg config) error {
	c, err := ensureConfigYandexMDB(cfg)
	if err != nil {
		log.Errorf("Error creating yandex discovery config: %v", err)
		return err
	}
	yd.config = c
	return nil
}

// Start implementation Start method of Discovery interface
func (yd *YandexDiscovery) Start(ctx context.Context, errCh chan<- error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for _, c := range yd.config {
		sdk, err := yandex.NewSDK(c.AuthorizedKey)
		if err != nil {
			log.Errorf("Error creating yandex sdk: %v", err)
			return err
		}
		var engine = yandexEngine{sdk: sdk, config: c, dsn: make(map[hostDb]clusterDSN)}
		err = engine.Start(ctx)
		if err != nil {
			log.Errorf("Error starting yandex engine: %v", err)
			return err
		}
		yd.engines = append(yd.engines, &engine)
	}

	for {
		err := yd.Sync()
		if err != nil {
			log.Errorf("YCD Sync error: %s", err.Error())
			errCh <- err
		}
		select {
		case <-ctx.Done():
			log.Debug("YCD context done")
			return nil
		default:
			time.Sleep(10 * time.Second)
		}
	}
}

func ensureConfigYandexMDB(config config) ([]YandexConfig, error) {
	c := &[]YandexConfig{}
	o, err := yaml.Marshal(config)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(o, c)
	if err != nil {
		return nil, err
	}
	return *c, nil
}
