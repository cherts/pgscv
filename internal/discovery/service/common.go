package service

import (
	"github.com/cherts/pgscv/discovery"
	"github.com/cherts/pgscv/discovery/log"
	"github.com/cherts/pgscv/internal/discovery/mapops"
)

// Label struct describe targets labels
type Label struct {
	Name  string `yaml:"name" json:"name"`
	Value string `yaml:"value" json:"value"`
}

// Env represents an environment variable configuration.
// It defines a name-value pair that will be set in the environment when running discovery.
type Env struct {
	// Name is the environment variable name. Must be a valid environment variable name
	// (starting with letter/underscore, containing only letters, numbers, and underscores).
	Name string `json:"name" yaml:"name" validate:"required,env_name"`
	// Value is the environment variable value that will.
	Value string `json:"value" yaml:"value" validate:"required"`
}

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
	if services == nil {
		log.Debugf("[%s SD] syncSubscriberServices: services is nil", provider)

		return nil
	}

	if subscribers == nil {
		log.Debugf("[%s SD] syncSubscriberServices: subscribers is nil", provider)

		return nil
	}

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
				targetLabels := make(map[string]string)
				labels["provider"] = provider

				if configLabels != nil {
					for _, l := range *configLabels {
						targetLabels[l.Name] = l.Value
					}
				}

				appendSvc[(*services)[*v.Left].name] = discovery.Service{
					DSN:          (*services)[*v.Left].dsn,
					ConstLabels:  labels,
					TargetLabels: targetLabels,
				}
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
