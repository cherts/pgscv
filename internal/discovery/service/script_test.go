package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cherts/pgscv/discovery"
	"github.com/cherts/pgscv/internal/discovery/script/response"
)

func TestFillSvcResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    *response.ScriptResponse
		setup    func()
		teardown func()
		want     string
		wantErr  bool
	}{
		{
			name: "empty DSN with all fields provided",
			input: &response.ScriptResponse{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
			},
			want: "postgres://testuser:testpass@localhost:5432/",
		},
		{
			name: "DSN provided, no overrides",
			input: &response.ScriptResponse{
				DSN: "postgres://user:pass@host:5432/dbname",
			},
			want: "postgres://user:pass@host:5432/dbname",
		},
		{
			name: "DSN provided with host override",
			input: &response.ScriptResponse{
				DSN:  "postgres://user:pass@oldhost:5432/dbname",
				Host: "newhost",
			},
			want: "postgres://user:pass@newhost:5432/dbname",
		},
		{
			name: "DSN provided with port override",
			input: &response.ScriptResponse{
				DSN:  "postgres://user:pass@host:5432/dbname",
				Port: 6432,
			},
			want: "postgres://user:pass@host:6432/dbname",
		},
		{
			name: "DSN provided with user override",
			input: &response.ScriptResponse{
				DSN:  "postgres://olduser:pass@host:5432/dbname",
				User: "newuser",
			},
			want: "postgres://newuser:pass@host:5432/dbname",
		},
		{
			name: "DSN provided with password override",
			input: &response.ScriptResponse{
				DSN:      "postgres://user:oldpass@host:5432/dbname",
				Password: "newpass",
			},
			want: "postgres://user:newpass@host:5432/dbname",
		},
		{
			name: "user from environment variable",
			input: &response.ScriptResponse{
				Host:        "localhost",
				Port:        5432,
				UserFromEnv: "TEST_USER_ENV",
			},
			setup: func() {
				_ = os.Setenv("TEST_USER_ENV", "envuser")
			},
			teardown: func() {
				_ = os.Unsetenv("TEST_USER_ENV")
			},
			want: "postgres://envuser@localhost:5432/",
		},
		{
			name: "password from environment variable",
			input: &response.ScriptResponse{
				Host:               "localhost",
				Port:               5432,
				User:               "testuser",
				PasswordFromEnvVar: "TEST_PASS_ENV",
			},
			setup: func() {
				_ = os.Setenv("TEST_PASS_ENV", "envpass")
			},
			teardown: func() {
				_ = os.Unsetenv("TEST_PASS_ENV")
			},
			want: "postgres://testuser:envpass@localhost:5432/",
		},
		{
			name: "environment variable not set, fallback to empty",
			input: &response.ScriptResponse{
				Host:        "localhost",
				Port:        5432,
				UserFromEnv: "NONEXISTENT_ENV",
			},
			want: "postgres://localhost:5432/",
		},
		{
			name: "invalid DSN should return error",
			input: &response.ScriptResponse{
				DSN: "invalid://dsn",
			},
			wantErr: true,
		},
		{
			name: "only host and port",
			input: &response.ScriptResponse{
				Host: "db.example.com",
				Port: 5432,
			},
			want: "postgres://db.example.com:5432/",
		},
		{
			name: "only user and password",
			input: &response.ScriptResponse{
				User:     "user",
				Password: "pass",
			},
			want: "postgres://user:pass/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if tt.teardown != nil {
				defer tt.teardown()
			}

			err := fillSvcResponse(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, tt.input.DSN)
			}
		})
	}
}

func TestFillSvcResponse_EdgeCases(t *testing.T) {
	t.Run("nil input should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = fillSvcResponse(nil)
		})
	})

	t.Run("empty input", func(t *testing.T) {
		svc := &response.ScriptResponse{}
		err := fillSvcResponse(svc)
		assert.NoError(t, err)
		assert.Equal(t, "postgres:///", svc.DSN)
	})

	t.Run("port 0 should be preserved", func(t *testing.T) {
		svc := &response.ScriptResponse{
			Host: "localhost",
			Port: 0,
		}
		err := fillSvcResponse(svc)
		assert.NoError(t, err)
		assert.Equal(t, "postgres://localhost/", svc.DSN)
	})
}

func TestFillSvcResponse_Priority(t *testing.T) {
	t.Run("user from field takes priority over env var", func(t *testing.T) {
		os.Setenv("TEST_USER", "envuser")
		defer os.Unsetenv("TEST_USER")

		svc := &response.ScriptResponse{
			Host:        "localhost",
			User:        "fielduser",
			UserFromEnv: "TEST_USER",
		}

		err := fillSvcResponse(svc)
		assert.NoError(t, err)
		assert.Equal(t, "postgres://fielduser@localhost/", svc.DSN)
	})

	t.Run("password from field takes priority over env var", func(t *testing.T) {
		os.Setenv("TEST_PASS", "envpass")
		defer os.Unsetenv("TEST_PASS")

		svc := &response.ScriptResponse{
			Host:               "localhost",
			User:               "user",
			Password:           "fieldpass",
			PasswordFromEnvVar: "TEST_PASS",
		}

		err := fillSvcResponse(svc)
		assert.NoError(t, err)
		assert.Equal(t, "postgres://user:fieldpass@localhost/", svc.DSN)
	})
}

const testScriptContent = `#!/bin/sh
echo "# service-id host port user password-from-env password"
echo "test-service-1 localhost 5432 postgres PG_PASSWORD -"
echo "test-service-2 192.168.1.100 6432 monitor - secret123"
`

func TestNewScriptDiscovery(t *testing.T) {
	sd := NewScriptDiscovery()
	assert.NotNil(t, sd)
	assert.NotNil(t, sd.subscribers)
	assert.Empty(t, sd.subscribers)
}

func TestScriptDiscovery_Init(t *testing.T) {
	sd := NewScriptDiscovery()

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test_script.sh")
	err := os.WriteFile(scriptPath, []byte(testScriptContent), 0755)
	require.NoError(t, err)

	config := scriptConfig{
		Script:           scriptPath,
		ExecutionTimeout: "10s",
		RefreshInterval:  "30s",
		OutputFormat:     "plain",
		Args:             []string{"--verbose"},
		Debug:            true,
	}

	err = sd.Init(config)
	assert.NoError(t, err)
	assert.Equal(t, scriptPath, sd.config.scriptPath)
	assert.Equal(t, 10*time.Second, sd.config.executionTimeoutDuration)
	assert.Equal(t, 30*time.Second, sd.config.refreshIntervalDuration)
	assert.True(t, sd.config.Debug)
}

func TestScriptDiscovery_Init_WithEnvVars(t *testing.T) {
	sd := NewScriptDiscovery()

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test_script.sh")
	err := os.WriteFile(scriptPath, []byte(testScriptContent), 0755)
	require.NoError(t, err)

	config := scriptConfig{
		Script:           scriptPath,
		ExecutionTimeout: "5s",
		RefreshInterval:  "15s",
		Env:              []Env{{Name: "TEST_VAR", Value: "test_value"}},
	}

	err = sd.Init(config)
	assert.NoError(t, err)
	require.Len(t, sd.config.Env, 1)
	assert.Equal(t, "TEST_VAR", sd.config.Env[0].Name)
	assert.Equal(t, "test_value", sd.config.Env[0].Value)
}

func TestScriptDiscovery_Subscribe_Unsubscribe(t *testing.T) {
	sd := NewScriptDiscovery()

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test_script.sh")
	err := os.WriteFile(scriptPath, []byte(testScriptContent), 0755)
	require.NoError(t, err)

	config := scriptConfig{
		Script:           scriptPath,
		ExecutionTimeout: "5s",
		RefreshInterval:  "10s",
	}

	err = sd.Init(config)
	require.NoError(t, err)

	subscriberID := "test-subscriber"

	addFunc := func(_ map[string]discovery.Service) error {
		return nil
	}

	removeFunc := func(_ []string) error {
		return nil
	}

	err = sd.Subscribe(subscriberID, addFunc, removeFunc)
	assert.NoError(t, err)

	sd.RLock()
	assert.Contains(t, sd.subscribers, subscriberID)
	sd.RUnlock()

	err = sd.Unsubscribe(subscriberID)
	assert.NoError(t, err)

	sd.RLock()
	assert.NotContains(t, sd.subscribers, subscriberID)
	sd.RUnlock()
}

func TestScriptDiscovery_getScriptResponse(t *testing.T) {
	sd := NewScriptDiscovery()

	scriptContent := `#!/bin/sh
echo "# service-id host port user password"
echo "cluster-1 db1.example.com 5432 postgres secret1"
echo "cluster-2 db2.example.com 6432 monitor secret2"
`

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test_script.sh")
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	config := scriptConfig{
		Script:           scriptPath,
		ExecutionTimeout: "5s",
		RefreshInterval:  "10s",
	}

	err = sd.Init(config)
	require.NoError(t, err)

	ctx := context.Background()
	responses, err := sd.getScriptResponse(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, responses)
	assert.Len(t, *responses, 2)

	expectedServices := []response.ScriptResponse{
		{
			ServiceID: "cluster-1",
			Host:      "db1.example.com",
			Port:      5432,
			User:      "postgres",
			Password:  "secret1",
		},
		{
			ServiceID: "cluster-2",
			Host:      "db2.example.com",
			Port:      6432,
			User:      "monitor",
			Password:  "secret2",
		},
	}

	for i, resp := range *responses {
		assert.Equal(t, expectedServices[i].ServiceID, resp.ServiceID)
		assert.Equal(t, expectedServices[i].Host, resp.Host)
		assert.Equal(t, expectedServices[i].Port, resp.Port)
		assert.Equal(t, expectedServices[i].User, resp.User)
		assert.Equal(t, expectedServices[i].Password, resp.Password)
	}
}

func TestScriptDiscovery_getScriptResponse_WithEnvVars(t *testing.T) {
	sd := NewScriptDiscovery()

	scriptContent := `#!/bin/sh
echo "# service-id host port user-from-env"
echo "test-cluster localhost 5432 PG_USER"
`

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test_script.sh")
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	envVarName := "PG_USER"
	envVarValue := "test_user"
	t.Setenv(envVarName, envVarValue)

	config := scriptConfig{
		Script:           scriptPath,
		ExecutionTimeout: "5s",
		RefreshInterval:  "15s",
		Env:              []Env{{Name: "TEST_VAR", Value: "test_value"}},
	}

	err = sd.Init(config)
	require.NoError(t, err)

	ctx := context.Background()
	responses, err := sd.getScriptResponse(ctx)
	assert.NoError(t, err)
	assert.Len(t, *responses, 1)

	resp := (*responses)[0]
	assert.Equal(t, "test-cluster", resp.ServiceID)
	assert.Equal(t, "localhost", resp.Host)
	assert.Equal(t, uint16(5432), resp.Port)
	assert.Equal(t, "PG_USER", resp.UserFromEnv)
}

func TestScriptDiscovery_getScriptResponse_InvalidOutput(t *testing.T) {
	sd := NewScriptDiscovery()

	scriptContent := `#!/bin/sh
echo "invalid output without header"
echo "just some random text"
`

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test_script.sh")
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	config := scriptConfig{
		Script:           scriptPath,
		ExecutionTimeout: "5s",
		RefreshInterval:  "15s",
	}

	err = sd.Init(config)
	require.NoError(t, err)

	ctx := context.Background()
	responses, err := sd.getScriptResponse(ctx)
	assert.NoError(t, err) // Should not error, just skip invalid lines
	assert.NotNil(t, responses)
	assert.Len(t, *responses, 0)
}

func TestScriptDiscovery_getServices(t *testing.T) {
	sd := NewScriptDiscovery()

	scriptContent := `#!/bin/sh
echo "# service-id host port user password-from-env password"
echo "main-cluster db-primary.com 5432 postgres PG_PASSWORD -"
echo "replica-cluster db-replica.com 5432 monitor - replica_pass"
echo "backup-cluster db-backup.com 6432 backup - backup_pass"
`

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test_script.sh")
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	config := scriptConfig{
		Script:           scriptPath,
		ExecutionTimeout: "5s",
		RefreshInterval:  "15s",
	}

	err = sd.Init(config)
	require.NoError(t, err)

	ctx := context.Background()
	services, err := sd.getServices(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, services)
	assert.Len(t, *services, 3)

	expectedServiceIDs := []string{"main-cluster", "replica-cluster", "backup-cluster"}
	for _, serviceID := range expectedServiceIDs {
		assert.Contains(t, *services, serviceID)
		clusterDSN := (*services)[serviceID]
		assert.Equal(t, serviceID, clusterDSN.name)
		assert.Contains(t, clusterDSN.dsn, "postgres://")
	}
}

func TestScriptDiscovery_getServices_WithEmptyResponse(t *testing.T) {
	sd := NewScriptDiscovery()

	scriptContent := `#!/bin/sh
echo "# service-id host port user password"
echo "- - - - -"
`

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test_script.sh")
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	config := scriptConfig{
		Script:           scriptPath,
		ExecutionTimeout: "5s",
		RefreshInterval:  "15s",
	}

	err = sd.Init(config)
	require.NoError(t, err)

	ctx := context.Background()
	services, err := sd.getServices(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, services)
	assert.Len(t, *services, 0) // Empty responses should be filtered out
}

func TestScriptDiscovery_Sync_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sd := NewScriptDiscovery()

	scriptContent := `#!/bin/sh
echo "# service-id host port user password"
echo "production-db db-prod.example.com 5432 monitor prod_password"
echo "staging-db db-staging.example.com 5432 monitor staging_password"
`

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test_script.sh")
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	config := scriptConfig{
		Script:           scriptPath,
		ExecutionTimeout: "5s",
		RefreshInterval:  "15s",
	}

	err = sd.Init(config)
	require.NoError(t, err)

	subscriberID := "test-sync-subscriber"
	var receivedServices map[string]discovery.Service

	addFunc := func(services map[string]discovery.Service) error {
		receivedServices = services
		return nil
	}

	removeFunc := func(_ []string) error {
		return nil
	}

	err = sd.Subscribe(subscriberID, addFunc, removeFunc)
	require.NoError(t, err)

	ctx := context.Background()
	err = sd.sync(ctx)
	assert.NoError(t, err)

	assert.NotNil(t, receivedServices)
	assert.Len(t, receivedServices, 2)

	for svcID, service := range receivedServices {
		assert.Contains(t, []string{"production-db", "staging-db"}, svcID)
		assert.Contains(t, service.DSN, "postgres://")
		assert.Contains(t, service.DSN, "monitor:")
	}
}

func TestScriptDiscovery_EnvironmentIsolation(t *testing.T) {
	sd := NewScriptDiscovery()

	scriptContent := `#!/bin/sh
echo "# service-id host"
echo "test-service localhost"
`

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test_script.sh")
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	originalValue := "original_value"
	t.Setenv("TEST_VAR", originalValue)

	config := scriptConfig{
		Script:           scriptPath,
		ExecutionTimeout: "5s",
		RefreshInterval:  "15s",
		Env:              []Env{{Name: "TEST_VAR", Value: "test_value"}},
	}

	err = sd.Init(config)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = sd.getServices(ctx)
	assert.NoError(t, err)

	assert.Equal(t, originalValue, os.Getenv("TEST_VAR"))
}

func TestScriptConfig_validate(t *testing.T) {
	tests := []struct {
		name    string
		config  scriptConfig
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: scriptConfig{
				Script:           "/bin/echo",
				ExecutionTimeout: "10s",
				RefreshInterval:  "30s",
			},
			wantErr: false,
		},
		{
			name: "missing script",
			config: scriptConfig{
				Script:           "",
				ExecutionTimeout: "10s",
				RefreshInterval:  "30s",
			},
			wantErr: true,
		},
		{
			name: "invalid execution timeout",
			config: scriptConfig{
				Script:           "/bin/echo",
				ExecutionTimeout: "invalid",
				RefreshInterval:  "30s",
			},
			wantErr: true,
		},
		{
			name: "invalid refresh interval",
			config: scriptConfig{
				Script:           "/bin/echo",
				ExecutionTimeout: "10s",
				RefreshInterval:  "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
