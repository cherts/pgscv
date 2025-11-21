package service

import (
	"time"

	"github.com/cherts/pgscv/internal/validators"
	"github.com/go-playground/validator/v10"
)

// scriptConfig holds the configuration for script-based service discovery.
type scriptConfig struct {
	// Script path to the discovery script
	Script string `json:"script" yaml:"script" validate:"required,regular_file"`
	// OutputFormat script output format (only plain supported)
	OutputFormat string `json:"output_format" yaml:"output_format" validate:"omitempty,oneof=plain"`
	// ExecutionTimeout maximum script execution time
	ExecutionTimeout string `json:"execution_timeout" yaml:"execution_timeout" validate:"required,ttl"`
	// Args command-line arguments for the script
	Args []string `json:"args" yaml:"args"`
	// Env environment variables to set for script execution
	Env []Env `json:"env" yaml:"env"`
	// RefreshInterval how often to run discovery @see time.ParseDuration()
	RefreshInterval string `json:"refresh_interval" yaml:"refresh_interval" validate:"required,ttl"`
	// Labels to apply to discovered services
	Labels *[]Label `json:"labels" yaml:"labels"`
	// TargetLabels target-specific labels
	TargetLabels *[]Label `json:"target_labels" yaml:"target_labels"`
	// Debug enable debug logging (insecure, dont use in production env)
	Debug bool `json:"debug" yaml:"debug"`
	// scriptPath cleaned script path
	scriptPath string
	// executionTimeoutDuration parsed execution timeout
	executionTimeoutDuration time.Duration
	// refreshIntervalDuration parsed refresh interval
	refreshIntervalDuration time.Duration
}

// validate performs validation on the script configuration using custom validators.
//
// Returns:
//   - error: if validation fails
func (c *scriptConfig) validate() error {
	validate := validator.New()

	registerCustomValidators(validate)

	err := validate.Struct(c)
	if err != nil {
		return err
	}

	return nil
}

// registerCustomValidators registers custom validation functions for script configuration.
//
// Parameters:
//   - v: validator instance to register custom validators with
func registerCustomValidators(v *validator.Validate) {
	_ = v.RegisterValidation(validators.TTLValidator, validators.TTLValidatorFunc)
	_ = v.RegisterValidation(validators.EnvNameValidator, validators.EnvNameValidatorFunc)
	_ = v.RegisterValidation(validators.RegularFileValidator, validators.RegularFileValidatorFunc)
}
