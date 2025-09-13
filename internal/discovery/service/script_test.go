package service

import (
	"github.com/cherts/pgscv/internal/discovery/script/response"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
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
