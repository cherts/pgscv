// Package service is main package of service discovery module
package service

import (
	"context"
	"fmt"
	"github.com/cherts/pgscv/internal/log"
	"gopkg.in/yaml.v2"
)

type config interface{}

// Service abstract service definition
type Service struct {
	DSN         string
	ConstLabels map[string]string
}

// AddServiceFunc services arg is map serviceId -> data source name
type AddServiceFunc func(services map[string]Service) error

// RemoveServiceFunc serviceIds is array of service IDs
type RemoveServiceFunc func(serviceIds []string) error

// Discovery interface of abstract discovery services
type Discovery interface {
	//Init casting abstract config to determined structure
	Init(c config) error
	//Start is the discoverer starting point
	Start(ctx context.Context, errCh chan<- error) error
	//Subscribe - binding "add" and "remove" functions. Functions will be called when services appear or disappear
	Subscribe(subscriberID string, addService AddServiceFunc, removeService RemoveServiceFunc) error
	//Unsubscribe - remove subscriber from list
	Unsubscribe(subscriberID string) error
}

const (
	yandexMDB = "yandex-mdb"
)

type sdConfig struct {
	Type   string `yaml:"type"`
	Config config `yaml:"config"`
}

// Instantiate returns initialized service discoverers (converting abstract configs to determined structures)
func Instantiate(discoveryConfig config) (*map[string]Discovery, error) {
	log.Debug("[Service Discovery] Instantiating")
	var config = make(map[string]sdConfig)
	var out, err = yaml.Marshal(discoveryConfig)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(out, config)
	if err != nil {
		return nil, err
	}
	var services = make(map[string]Discovery)
	for id, srv := range config {
		log.Debugf("[Service Discovery] found config %s", srv.Type)
		switch srv.Type {
		case yandexMDB:
			services[id] = &YandexDiscovery{subscribers: make(map[string]subscriber)}
		default:
			err := fmt.Errorf("[Service Discovery] unknown service discovery type: %s", srv.Type)
			log.Debug(err.Error())
			return nil, err
		}
		err := services[id].Init(srv.Config)
		if err != nil {
			log.Errorf("[Service Discovery] Error initializing %s: %s", id, err.Error())
			return nil, err
		}
	}
	return &services, nil
}
