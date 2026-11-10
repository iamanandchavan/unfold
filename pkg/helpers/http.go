package helpers

import (
	"context"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
)

// HTTPTimeoutEnv names the environment variable overriding the per-request HTTP timeout, e.g. "45s".
const HTTPTimeoutEnv = "UNFOLD_HTTP_TIMEOUT"

const defaultHTTPTimeout = 30 * time.Second

// HTTPTimeout returns the per-request timeout for calls to cloud provider APIs:
// the value of UNFOLD_HTTP_TIMEOUT when it is a positive duration, 30s otherwise.
func HTTPTimeout() time.Duration {
	if d, err := time.ParseDuration(os.Getenv(HTTPTimeoutEnv)); err == nil && d > 0 {
		return d
	}
	return defaultHTTPTimeout
}

// OAuthContext returns a context whose OAuth2 token requests are bounded by HTTPTimeout.
func OAuthContext() context.Context {
	return context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{Timeout: HTTPTimeout()})
}

// NewOAuthClient returns an HTTP client authorised by ts whose requests are bounded by HTTPTimeout.
func NewOAuthClient(ctx context.Context, ts oauth2.TokenSource) *http.Client {
	client := oauth2.NewClient(ctx, ts)
	client.Timeout = HTTPTimeout()
	// Hide oauth2.Transport's deprecated CancelRequest (it only logs a warning)
	// so the client cancels timed-out requests through the request context.
	client.Transport = struct{ http.RoundTripper }{client.Transport}
	return client
}
