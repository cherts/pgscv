package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/cherts/pgscv/discovery"
	"github.com/cherts/pgscv/discovery/log"
	"github.com/cherts/pgscv/internal/discovery/script/response"
	"github.com/jackc/pgx/v5"
	"gopkg.in/yaml.v2"
	"maps"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ScriptDiscovery is main struct for Script discoverer
type ScriptDiscovery struct {
	sync.RWMutex
	config      scriptConfig
	subscribers map[string]subscriber
}

func NewScriptDiscovery() *ScriptDiscovery {
	return &ScriptDiscovery{subscribers: make(map[string]subscriber)}
}

// Init implementation Init method of Discovery interface
func (s *ScriptDiscovery) Init(c discovery.Config) error {
	log.Debug("[Script SD] Init discovery config...")

	pc, err := ensureConfigScript(c)
	if err != nil {
		log.Errorf("[Script SD] Failed to init discovery config, error: %v", err)
		return err
	}

	s.config = *pc
	return nil
}

// Start implementation Start method of Discovery interface
func (s *ScriptDiscovery) Start(ctx context.Context, errCh chan<- error) error {
	s.RLock()
	refreshInterval := s.config.refreshIntervalDuration
	s.RUnlock()

	for {
		err := s.sync(ctx)
		if err != nil {
			log.Errorf("[Script SD] Failed to sync, error: %s", err.Error())
			errCh <- err
		}

		select {
		case <-ctx.Done():
			log.Debug("[Script SD] Context done.")

			return nil
		default:
			time.Sleep(refreshInterval)
		}
	}
}

// Subscribe implementation Subscribe method of Discovery interface
func (s *ScriptDiscovery) Subscribe(subscriberID string, addService discovery.AddServiceFunc, removeService discovery.RemoveServiceFunc) error {
	s.Lock()
	defer s.Unlock()

	s.subscribers[subscriberID] = subscriber{
		AddService:     addService,
		RemoveService:  removeService,
		syncedServices: make(map[string]discovery.Service),
		SyncedVersion:  make(map[engineIdx]version),
	}

	return nil
}

// Unsubscribe implementation Unsubscribe method of Discovery interface
func (s *ScriptDiscovery) Unsubscribe(subscriberID string) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.subscribers[subscriberID]; !ok {
		return nil
	}

	svc := make([]string, 0, len(s.subscribers[subscriberID].syncedServices))

	for k := range maps.Keys(s.subscribers[subscriberID].syncedServices) {
		svc = append(svc, k)
	}

	err := s.subscribers[subscriberID].RemoveService(svc)

	delete(s.subscribers, subscriberID)

	return err
}

func (s *ScriptDiscovery) sync(ctx context.Context) error {
	s.Lock()
	defer s.Unlock()

	services, err := s.getServices(ctx)
	if err != nil {
		log.Errorf("[Script SD] Failed to get services, error: %v", err)
		return nil // non fatal, scripts may fail
	}

	err = syncSubscriberServices(discovery.Script, &s.subscribers, services, s.config.TargetLabels)
	if err != nil {
		return err
	}

	return nil
}

// preserve environment consistency
var envLock sync.Mutex

func (s *ScriptDiscovery) getServices(ctx context.Context) (*map[string]clusterDSN, error) {
	envLock.Lock()

	oldEnvValues, err := s.oldEnvValues()
	if err != nil {
		return nil, err
	}

	defer func() {
		s.restoreEnv(oldEnvValues)
		envLock.Unlock()
	}()

	services, err := s.getScriptResponse(ctx)
	if err != nil {
		return nil, err
	}

	ret := map[string]clusterDSN{}

	for n, svc := range *services {
		if svc.AllFieldsEmpty() {
			continue
		}

		err = svc.Validate()
		if err != nil {
			log.Errorf("[Script SD] Failed to validate svc config #%d, error: %v", n, err)

			continue
		}

		err = fillSvcResponse(&svc)
		if err != nil {
			log.Errorf("[Script SD] Failed to fill svc config #%d, error: %v", n, err)

			continue
		}

		ret[svc.ServiceID] = clusterDSN{name: svc.ServiceID, dsn: svc.DSN}
	}

	return &ret, nil
}

func fillSvcResponse(svc *response.ScriptResponse) error {
	var (
		dbConfig *pgx.ConnConfig
		err      error
	)

	if svc.DSN == "" {
		dbConfig = &pgx.ConnConfig{}
	} else {
		dbConfig, err = pgx.ParseConfig(svc.DSN)
		if err != nil {
			return err
		}
	}

	if svc.Host != "" {
		dbConfig.Host = svc.Host
	}

	if svc.Port != 0 {
		dbConfig.Port = svc.Port
	}

	if svc.User != "" {
		dbConfig.User = svc.User
	} else if svc.UserFromEnv != "" {
		if envVal, exists := os.LookupEnv(svc.UserFromEnv); exists {
			dbConfig.User = envVal
		}
	}

	if svc.Password != "" {
		dbConfig.Password = svc.Password
	} else if svc.PasswordFromEnvVar != "" {
		if envVal, exists := os.LookupEnv(svc.PasswordFromEnvVar); exists {
			dbConfig.Password = envVal
		}
	}

	var (
		hostPort         string
		userPass         string
		userPassSlice    = make([]string, 0)
		credentialsSlice = make([]string, 0)
	)

	if dbConfig.User != "" {
		userPassSlice = append(userPassSlice, dbConfig.User)
		if dbConfig.Password != "" {
			userPassSlice = append(userPassSlice, dbConfig.Password)
		}
		userPass = strings.Join(userPassSlice, ":")
		credentialsSlice = append(credentialsSlice, userPass)
	}

	if dbConfig.Host != "" {
		if dbConfig.Port > 0 {
			hostPort = net.JoinHostPort(dbConfig.Host, strconv.Itoa(int(dbConfig.Port)))
		} else {
			hostPort = dbConfig.Host
		}
		credentialsSlice = append(credentialsSlice, hostPort)
	}

	credentials := strings.Join(credentialsSlice, "@")

	svc.DSN = fmt.Sprintf("postgres://%s/%s", credentials, dbConfig.Database)

	return nil
}

func (s *ScriptDiscovery) getScriptResponse(ctx context.Context) (*[]response.ScriptResponse, error) {
	execCtx, execCancel := context.WithTimeout(ctx, s.config.executionTimeoutDuration)
	defer execCancel()

	cmd := exec.CommandContext(execCtx, s.config.scriptPath, s.config.Args...)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	stdout, exErr := cmd.Output()
	if exErr != nil {
		var exitErr *exec.ExitError

		ok := errors.As(exErr, &exitErr)
		if ok {
			exitCode := exitErr.ExitCode()

			return nil, fmt.Errorf("exit_code: %d stderr: %s, err: %w", exitCode, stderr.String(), exitErr)
		}

		return nil, exErr
	}

	commandOutputString := string(stdout)
	stdErrOutput := stderr.String()

	if stdErrOutput != "" {
		log.Debugf("[Script SD] Command stderr output: %s", stdErrOutput)
	}

	if s.config.Debug {
		log.Debugf("[Script SD] Command output: %s", commandOutputString)
	}

	services := &[]response.ScriptResponse{}

	err := response.UnmarshalScriptResponse(commandOutputString, services)
	if err != nil {
		return nil, err
	}

	return services, nil
}

func (s *ScriptDiscovery) restoreEnv(oldEnvValues map[string]*string) {
	for envName, oldEnvValue := range oldEnvValues {
		if oldEnvValue == nil {
			err := os.Unsetenv(envName)

			if err != nil {
				log.Errorf("[Script SD] Failed to unset environment variable %s: %v", envName, err)
			}
		} else {
			err := os.Setenv(envName, *oldEnvValue)
			if err != nil {
				log.Errorf("[Script SD] Failed to set environment variable %s: %v", envName, err)
			}
		}
	}
}

func (s *ScriptDiscovery) oldEnvValues() (map[string]*string, error) {
	oldEnvValues := make(map[string]*string, len(s.config.Env))
	for _, env := range s.config.Env {
		envVar, present := os.LookupEnv(env.Name)
		if !present {
			oldEnvValues[env.Name] = nil
		} else {
			oldEnvValues[env.Name] = &envVar
		}

		err := os.Setenv(env.Name, envVar)
		if err != nil {
			return nil, err
		}
	}

	return oldEnvValues, nil
}

func ensureConfigScript(config discovery.Config) (*scriptConfig, error) {
	c := &scriptConfig{}

	o, err := yaml.Marshal(config)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(o, c)
	if err != nil {
		return nil, err
	}

	err = c.validate()
	if err != nil {
		return nil, err
	}

	c.scriptPath = filepath.Clean(c.Script)
	// errors checked in ttl validator
	c.executionTimeoutDuration, _ = time.ParseDuration(c.ExecutionTimeout)
	c.refreshIntervalDuration, _ = time.ParseDuration(c.RefreshInterval)

	return c, nil
}
