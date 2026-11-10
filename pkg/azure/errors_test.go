package azure

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func Test_tenantLookupTransportError(t *testing.T) {
	prepareConfig()
	prepareTestEnvironment()
	instances[managementResourceIndex].httpClient = &http.Client{
		Transport: &MockHTTPRoundTripper{Error: errors.New("dial tcp: connection refused")},
	}

	tenants, err := NewTenantFinder().GetTenantBySubscriptionID("12345678-1234-1234-1234-123456789abc")

	if err == nil || !strings.Contains(err.Error(), "connection refused") {
		t.Fatalf("want transport error, got tenants=%v err=%v", tenants, err)
	}
}

func Test_getPlansErrorKeepsResponseBody(t *testing.T) {
	prepareConfig()
	prepareTestEnvironment()
	const plansURL = "https://graph.microsoft.com/rp/product-ingestion/plan?product=product/87654321-4321-4321-4321-210987654321&$version=2022-03-01-preview2"
	instances[graphResourceIndex].httpClient = &http.Client{
		Transport: &MockHTTPRoundTripper{Transport: map[string]*http.Response{
			plansURL: {
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error": {"code": "BadRequest", "message": "product not found"}}`)),
			},
		}},
	}

	got := MakeConfigurationRequest("offer-2", "12345678-1234-1234-1234-123456789abc", "tenant", AddMode)

	if !strings.Contains(got, "product not found") {
		t.Errorf("error lost the response body: %q", got)
	}
}
