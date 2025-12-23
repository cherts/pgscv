package yandex

import (
	"encoding/json"
	"fmt"
	"github.com/cherts/pgscv/discovery/log"
	"github.com/go-playground/validator/v10"
	"os"
	"path/filepath"
	"time"
)

type authorizedKey struct {
	ID               string    `json:"id"`
	ServiceAccountID string    `json:"service_account_id" validate:"required"`
	CreatedAt        time.Time `json:"created_at" validate:"required"`
	KeyAlgorithm     string    `json:"key_algorithm" validate:"required"`
	PublicKey        string    `json:"public_key" validate:"required"`
	PrivateKey       string    `json:"private_key" validate:"required"`
}

func (k *authorizedKey) validate() error {
	v := validator.New()

	err := v.Struct(k)
	if err != nil {
		return fmt.Errorf("validate | %w", err)
	}

	return nil
}

func loadAuthorizedKey(filePath string) (*authorizedKey, error) {

	log.Debugf("[SD] Loading authorized key from path '%s'", filePath)
	data, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		log.Errorf("[SD] Failed to load authorized key, error: %s", err.Error())
		return nil, err
	}

	var key authorizedKey

	err = json.Unmarshal(data, &key)
	if err != nil {
		log.Errorf("[SD] Failed to parse authorized key, JSON parse error: %s", err.Error())
		return nil, err
	}

	err = key.validate()
	if err != nil {
		log.Errorf("[SD] Failed to validate authorized key, error: %s", err.Error())
	}

	return &key, nil
}
