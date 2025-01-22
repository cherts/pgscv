package service

import "testing"

func TestInstantiate(t *testing.T) {

	testCases := []struct {
		name    string
		cfg     map[string]sdConfig
		wantErr bool
	}{
		{
			name: "succeed single service",
			cfg: map[string]sdConfig{
				"yandex1": {
					Type: yandexMDB,
					Config: []YandexConfig{
						{
							AuthorizedKey: "/tmp/authorized_key.json",
							FolderID:      "asd234234234",
							User:          "postgres_exporter",
							Password:      "132",
							Clusters: []cluster{
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
			cfg: map[string]sdConfig{
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
			s, err := Instantiate(tc.cfg)
			if (err != nil) != tc.wantErr {
				t.Errorf("Instantiate() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			for _, v := range *s {
				err := v.Subscribe("svc",
					func(_ map[string]Service) error {
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
