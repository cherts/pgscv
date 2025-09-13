package service

import (
	"github.com/cherts/pgscv/internal/validators"
	"github.com/go-playground/validator/v10"
	"time"
)

type scriptConfig struct {
	Script                   string   `json:"script" yaml:"script" validate:"required,regular_file"`
	OutputFormat             string   `json:"output_format" yaml:"output_format" validate:"omitempty,oneof=plain"`
	ExecutionTimeout         string   `json:"execution_timeout" yaml:"execution_timeout" validate:"required,ttl"`
	Args                     []string `json:"args" yaml:"args"`
	Env                      []Env    `json:"env" yaml:"env"`
	RefreshInterval          string   `json:"refresh_interval" yaml:"refresh_interval" validate:"required,ttl"`
	Labels                   *[]Label `json:"labels" yaml:"labels"`
	TargetLabels             *[]Label `json:"target_labels" yaml:"target_labels"`
	Debug                    bool     `json:"debug" yaml:"debug"`
	scriptPath               string
	executionTimeoutDuration time.Duration
	refreshIntervalDuration  time.Duration
}

func (c *scriptConfig) validate() error {
	validate := validator.New()

	registerCustomValidators(validate)

	err := validate.Struct(c)
	if err != nil {
		return err
	}

	return nil
}

func registerCustomValidators(v *validator.Validate) {
	_ = v.RegisterValidation(validators.TTLValidator, validators.TTLValidatorFunc)
	_ = v.RegisterValidation(validators.EnvNameValidator, validators.EnvNameValidatorFunc)
	_ = v.RegisterValidation(validators.RegularFileValidator, validators.RegularFileValidatorFunc)
}
