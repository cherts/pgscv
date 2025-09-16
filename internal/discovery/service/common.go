package service

import (
	"github.com/cherts/pgscv/discovery"
	"github.com/cherts/pgscv/discovery/log"
	"github.com/cherts/pgscv/internal/discovery/mapops"
)

type clusterDSN struct {
	dsn, name string
	labels    []Label
}

type subscriber struct {
	AddService     discovery.AddServiceFunc
	RemoveService  discovery.RemoveServiceFunc
	SyncedVersion  map[engineIdx]version
	syncedServices map[string]discovery.Service
}

func syncSubscriberServices(
	provider string,
	subscribers *map[string]subscriber,
	services *map[string]clusterDSN,
	configLabels *[]Label,
) error {
	for _, subscriber := range *subscribers {
		removeSvc := make([]string, 0)
		appendSvc := make(map[string]discovery.Service)
		for _, v := range mapops.FullJoin(*services, subscriber.syncedServices) {
			if v.Left == nil {
				removeSvc = append(removeSvc, *v.Right)
				delete(subscriber.syncedServices, *v.Right)
			}
			if v.Right == nil {
				labels := make(map[string]string)
				labels["provider"] = provider
				targetLabels := make(map[string]string)
				if configLabels != nil {
					for _, l := range *configLabels {
						targetLabels[l.Name] = l.Value
					}
				}
				appendSvc[(*services)[*v.Left].name] = discovery.Service{DSN: (*services)[*v.Left].dsn, ConstLabels: labels, TargetLabels: targetLabels}
				subscriber.syncedServices[*v.Left] = appendSvc[*v.Left]
			}
		}
		if len(removeSvc) > 0 {
			log.Debugf("[%s SD] Removing '%d' services from subscriber.", provider, len(removeSvc))
			err := subscriber.RemoveService(removeSvc)
			if err != nil {
				return err
			}
		}
		if len(appendSvc) > 0 {
			log.Debugf("[%s Discovery] Appending '%d' services to subscriber.", provider, len(appendSvc))
			err := subscriber.AddService(appendSvc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
