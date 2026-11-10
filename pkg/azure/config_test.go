package azure

import (
	"os"
	"testing"
	"time"
)

func TestStartService(t *testing.T) {
	tests := []struct {
		name         string
		modifyConfig func()
		wantErr      bool
	}{
		{
			name: "start service success",
			modifyConfig: func() {
				prepareConfig()
			},
			wantErr: false,
		},
		{
			name: "start service error no offers file",
			modifyConfig: func() {
				prepareConfig()
				os.Setenv("AZURE_OFFERS_FILE", "offers_test.yml")
			},
			wantErr: true,
		},
		{
			name: "start service error invalid offers file",
			modifyConfig: func() {
				prepareConfig()
				os.Setenv("AZURE_OFFERS_FILE", "testdata/invalid_offers_test.yml")
			},
			wantErr: true,
		},
		{
			name: "start service error cert file not found",
			modifyConfig: func() {
				prepareConfig()
				os.Setenv("AZURE_CERT_FILE", "testdata/ca_cert.txt")
			},
			wantErr: true,
		},
		{
			name: "start service error invalid cert file",
			modifyConfig: func() {
				prepareConfig()
				os.Setenv("AZURE_CERT_FILE", "testdata/invalid_ca_cert.txt")
			},
			wantErr: true,
		},
		{
			name: "start service error resource index out of bounds",
			modifyConfig: func() {
				prepareConfig()
				// add a new resource index in config, as iota cannot be modified
				instances[len(config.Resources)+1] = nil
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.modifyConfig != nil {
				tt.modifyConfig()
			}
			if err := StartService(); (err != nil) != tt.wantErr {
				t.Errorf("StartService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func prepareConfig() {
	os.Setenv("AZURE_OFFERS_FILE", "testdata/offers_test.yml")
	os.Setenv("AZURE_CERT_FILE", "testdata/test_ca_cert.txt")
	config.IdentityCAFile = os.Getenv("AZURE_CERT_FILE")
	instances = map[int]*AZService{
		managementResourceIndex: nil,
		graphResourceIndex:      nil,
	}
}

func TestStartServiceSetsHTTPTimeout(t *testing.T) {
	t.Setenv("UNFOLD_HTTP_TIMEOUT", "7s")
	prepareConfig()
	if err := StartService(); err != nil {
		t.Fatal(err)
	}
	for idx, svc := range instances {
		if svc.httpClient.Timeout != 7*time.Second {
			t.Errorf("instance %d timeout = %v, want 7s", idx, svc.httpClient.Timeout)
		}
	}
}
