package helpers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func TestHTTPTimeout(t *testing.T) {
	tests := []struct {
		env  string
		want time.Duration
	}{
		{"", 30 * time.Second},
		{"45s", 45 * time.Second},
		{"250ms", 250 * time.Millisecond},
		{"not-a-duration", 30 * time.Second},
		{"0s", 30 * time.Second},
		{"-5s", 30 * time.Second},
	}
	for _, tt := range tests {
		t.Setenv(HTTPTimeoutEnv, tt.env)
		if got := HTTPTimeout(); got != tt.want {
			t.Errorf("HTTPTimeout() with %q = %v, want %v", tt.env, got, tt.want)
		}
	}
}

// slowServer never answers until the client gives up or the test ends.
func slowServer(t *testing.T) *httptest.Server {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(release) }) // runs first: unblock handlers before Close waits on them
	return srv
}

func TestNewOAuthClientTimesOutAPICalls(t *testing.T) {
	t.Setenv(HTTPTimeoutEnv, "100ms")
	srv := slowServer(t)
	client := NewOAuthClient(OAuthContext(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "t"}))

	start := time.Now()
	resp, err := client.Get(srv.URL)
	if err == nil {
		resp.Body.Close()
		t.Fatal("want timeout error, got response")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("request took %v, want it bounded by the 100ms timeout", elapsed)
	}
}

func TestOAuthContextTimesOutTokenRequests(t *testing.T) {
	t.Setenv(HTTPTimeoutEnv, "100ms")
	tokenSrv := slowServer(t)
	conf := &clientcredentials.Config{ClientID: "id", ClientSecret: "secret", TokenURL: tokenSrv.URL}

	start := time.Now()
	if _, err := conf.TokenSource(OAuthContext()).Token(); err == nil {
		t.Fatal("want timeout error from token endpoint")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("token request took %v, want it bounded by the 100ms timeout", elapsed)
	}
}

func TestNewOAuthClientDoesNotUseDeprecatedCancelRequest(t *testing.T) {
	client := NewOAuthClient(OAuthContext(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "t"}))
	if _, ok := client.Transport.(interface{ CancelRequest(*http.Request) }); ok {
		t.Error("client transport exposes CancelRequest; oauth2 logs a deprecation warning when a timeout cancels a request")
	}
}
