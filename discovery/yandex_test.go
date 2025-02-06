package discovery_test

import (
	"github.com/cherts/pgscv/discovery"
	"github.com/cherts/pgscv/discovery/instantiate"
	"github.com/cherts/pgscv/internal/discovery/service"
	"testing"
)

func TestInstantiate(t *testing.T) {

	testCases := []struct {
		name    string
		cfg     map[string]discovery.SdConfig
		wantErr bool
	}{
		{
			name: "succeed single service",
			cfg: map[string]discovery.SdConfig{
				"yandex1": {
					Type: discovery.YandexMDB,
					Config: []service.YandexConfig{
						{
							AuthorizedKey: "/tmp/authorized_key.json",
							FolderID:      "asd234234234",
							User:          "postgres_exporter",
							Password:      "132",
							Clusters: []service.Cluster{
								{
									Db:        stringPtr(".*"),
									ExcludeDb: stringPtr("(postgres|template)"),
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "wrong single service",
			cfg: map[string]discovery.SdConfig{
				"yandex1": {
					Type:   "abcdefg",
					Config: []string{},
				},
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := instantiate.Instantiate(tc.cfg)
			if (err != nil) != tc.wantErr {
				t.Errorf("Instantiate() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			for _, v := range *s {
				err := v.Subscribe("svc",
					func(_ map[string]discovery.Service) error {
						return nil
					},
					func(_ []string) error {
						return nil
					},
				)
				if err != nil {
					t.Errorf("Subscribe() error = %v", err)
				}
				err = v.Unsubscribe("svc")
				if err != nil {
					t.Errorf("Unsubscribe() error = %v", err)
				}
			}
		})

	}

}

func stringPtr(s string) *string {
	return &s
}
