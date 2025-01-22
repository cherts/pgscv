// Package yandex implements getter for YC MDB PostgreSQL clusters, hosts and databases
package yandex

import (
	"context"
	"sync"

	"github.com/cherts/pgscv/internal/log"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

// SDK struct with limited lifespan token
type SDK struct {
	sync.RWMutex
	token *tokenIAM
}

// NewSDK load authorized key from json file and return pointer on SDK structure
func NewSDK(jsonFilePath string) (*SDK, error) {
	log.Debug("[SD] Loading authorized key from json file...")
	token, err := newIAMToken(jsonFilePath)
	if err != nil {
		return nil, err
	}
	return &SDK{
		token: token,
	}, nil
}

// Build creates an SDK instance
func (sdk *SDK) Build(ctx context.Context) (*ycsdk.SDK, error) {
	var t, err = sdk.token.GetToken()
	if err != nil {
		return nil, err
	}
	log.Debug("[SD] Build Yandex.Cloud SDK...")
	ysdk, err := ycsdk.Build(ctx, ycsdk.Config{
		Credentials: ycsdk.NewIAMTokenCredentials(*t),
	})
	if err != nil {
		log.Errorf("[SD] Failed to build Yandex.Cloud SDK, error: %v", err)
		return nil, err
	}
	return ysdk, nil
}
