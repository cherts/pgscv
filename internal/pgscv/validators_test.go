package pgscv

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestPoolConfigValidation(t *testing.T) {
	v := validator.New()
	registerCustomValidators(v)

	tests := []struct {
		name     string
		config   *Config
		expected bool
	}{
		{
			name:     "nil pool config should be valid",
			config:   &Config{PoolerConfig: nil},
			expected: true,
		},
		{
			name: "empty pool config should be valid",
			config: &Config{
				PoolerConfig: &PoolConfig{
					MaxConns:     nil,
					MinConns:     nil,
					MinIdleConns: nil,
				},
			},
			expected: true,
		},
		{
			name: "valid positive values",
			config: &Config{
				PoolerConfig: &PoolConfig{
					MaxConns:     new(int32(10)),
					MinConns:     new(int32(5)),
					MinIdleConns: new(int32(3)),
				},
			},
			expected: true,
		},
		{
			name: "max conns zero should be invalid",
			config: &Config{
				PoolerConfig: &PoolConfig{
					MaxConns: new(int32(0)),
				},
			},
			expected: false,
		},
		{
			name: "max conns negative should be invalid",
			config: &Config{
				PoolerConfig: &PoolConfig{
					MaxConns: new(int32(-5)),
				},
			},
			expected: false,
		},
		{
			name: "min conns greater than max conns should be invalid",
			config: &Config{
				PoolerConfig: &PoolConfig{
					MaxConns: new(int32(5)),
					MinConns: new(int32(10)),
				},
			},
			expected: false,
		},
		{
			name: "min idle conns greater than max conns should be invalid",
			config: &Config{
				PoolerConfig: &PoolConfig{
					MaxConns:     new(int32(5)),
					MinIdleConns: new(int32(10)),
				},
			},
			expected: false,
		},
		{
			name: "min idle conns greater than min conns should be invalid",
			config: &Config{
				PoolerConfig: &PoolConfig{
					MinConns:     new(int32(5)),
					MinIdleConns: new(int32(10)),
				},
			},
			expected: false,
		},
		{
			name: "all valid relationships",
			config: &Config{
				PoolerConfig: &PoolConfig{
					MaxConns:     new(int32(20)),
					MinConns:     new(int32(10)),
					MinIdleConns: new(int32(5)),
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(tt.config)

			if tt.expected {
				assert.NoError(t, err, "Expected config to be valid, but got error: %v", err)
			} else {
				assert.Error(t, err, "Expected config to be invalid, but validation passed")

				if err != nil {
					validationErrors, ok := err.(validator.ValidationErrors)
					assert.True(t, ok, "Error should be of type ValidationErrors")

					hasPoolConfigError := false
					for _, fieldError := range validationErrors {
						if fieldError.StructField() == "PoolerConfig" {
							hasPoolConfigError = true
							break
						}
					}
					assert.True(t, hasPoolConfigError, "Should have PoolerConfig validation error")
				}
			}
		})
	}
}

func TestPoolConfigFieldLevelValidation(t *testing.T) {
	v := validator.New()
	registerCustomValidators(v)

	tests := []struct {
		name     string
		poolCfg  *PoolConfig
		expected bool
	}{
		{
			name:     "nil config",
			poolCfg:  nil,
			expected: true,
		},
		{
			name: "valid individual field validation",
			poolCfg: &PoolConfig{
				MaxConns:     new(int32(10)),
				MinConns:     new(int32(5)),
				MinIdleConns: new(int32(2)),
			},
			expected: true,
		},
		{
			name: "invalid max conns",
			poolCfg: &PoolConfig{
				MaxConns: new(int32(-1)),
			},
			expected: false,
		},
		{
			name: "invalid min idle conns",
			poolCfg: &PoolConfig{
				MinIdleConns: new(int32(-5)),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{PoolerConfig: tt.poolCfg}
			err := v.Struct(config)

			if tt.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
