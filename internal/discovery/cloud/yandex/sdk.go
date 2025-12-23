// Package yandex implements getter for YC MDB PostgreSQL clusters, hosts and databases
package yandex

import (
	"context"
	"fmt"
	"sync"

	"github.com/cherts/pgscv/discovery/log"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SDK struct with limited lifespan token
type SDK struct {
	sync.RWMutex
	key   *authorizedKey
	ycsdk *ycsdk.SDK
}

// NewSDK load authorized key from json file and return pointer on SDK structure
func NewSDK(jsonFilePath string) (*SDK, error) {
	log.Debug("[SD] Loading authorized key from json file...")

	key, err := loadAuthorizedKey(jsonFilePath)
	if err != nil {
		return nil, err
	}
	return &SDK{
		key: key,
	}, nil
}

// Build creates an SDK instance
func (sdk *SDK) buildClient(ctx context.Context) (*ycsdk.SDK, error) {
	credentials, err := ycsdk.ServiceAccountKey(
		&iamkey.Key{ //nolint: exhaustruct
			Id:           sdk.key.ID,
			CreatedAt:    timestamppb.New(sdk.key.CreatedAt),
			KeyAlgorithm: iam.Key_Algorithm(iam.Key_Algorithm_value[sdk.key.KeyAlgorithm]),
			PublicKey:    sdk.key.PublicKey,
			PrivateKey:   sdk.key.PrivateKey,
			Subject:      &iamkey.Key_ServiceAccountId{ServiceAccountId: sdk.key.ServiceAccountID},
		})
	if err != nil {
		return nil, fmt.Errorf("[SD] failed to construct key | %w", err)
	}

	y, err := ycsdk.Build(ctx,
		ycsdk.Config{ //nolint: exhaustruct
			Credentials: credentials,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("[SD] failed to Build yandex sdk | %w", err)
	}

	return y, nil
}

// Build creates an SDK instance
func (sdk *SDK) Build(ctx context.Context) (*ycsdk.SDK, error) {
	var err error

	if sdk.ycsdk == nil {
		sdk.ycsdk, err = sdk.buildClient(ctx)

		if err != nil {
			return nil, err
		}
	}

	return sdk.ycsdk, nil
}
