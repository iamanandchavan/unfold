package google

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// stallingServer never answers until the client gives up or the test ends.
func stallingServer(t *testing.T) *httptest.Server {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(release) })
	return srv
}

// writeKeyfile writes a service account keyfile with a real RSA key and the given token endpoint.
func writeKeyfile(t *testing.T, tokenURI string) string {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(map[string]string{
		"type":         "service_account",
		"private_key":  string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})),
		"client_email": "sa@example.iam.gserviceaccount.com",
		"token_uri":    tokenURI,
	})
	path := filepath.Join(t.TempDir(), "keyfile.json")
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func Test_requestsAreBoundedByHTTPTimeout(t *testing.T) {
	tokenOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token": "t", "token_type": "Bearer", "expires_in": 3600}`))
	}))
	t.Cleanup(tokenOK.Close)
	stalled := stallingServer(t)

	tests := []struct {
		name, tokenURI, apiBase string
	}{
		{"token request stalls", stalled.URL, ""},
		{"api request stalls", tokenOK.URL, stalled.URL + "/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("UNFOLD_HTTP_TIMEOUT", "100ms")
			t.Setenv("GOOGLE_KEYFILE", writeKeyfile(t, tt.tokenURI))
			Config.clientOpts = nil
			if err := StartService(); err != nil {
				t.Fatal(err)
			}
			if tt.apiBase != "" {
				instance.CloudIdentityService.BasePath = tt.apiBase
			}

			start := time.Now()
			_, err := instance.CloudIdentityService.Groups.Lookup().GroupKeyId("g@example.com").Do()
			if err == nil {
				t.Fatal("want timeout error")
			}
			if elapsed := time.Since(start); elapsed > 2*time.Second {
				t.Errorf("call took %v, want it bounded by the 100ms timeout (err: %v)", elapsed, err)
			}
		})
	}
}
