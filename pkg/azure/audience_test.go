package azure

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

func Test_SearchAcrossPlans(t *testing.T) {
	prepareConfig()
	prepareTestEnvironment()
	const (
		product = "12345678-1234-1234-1234-123456789abc"
		treeURL = "https://graph.microsoft.com/rp/product-ingestion/resource-tree/product/" + product
		planA   = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
		planB   = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	)
	tests := []struct {
		name, tree, id, want string
	}{
		{
			name: "id present only in a later plan",
			tree: `{"resources": [
				{"$schema": "product"},
				{"privateAudiences": [{"id": "` + planA + `", "type": "tenant"}]},
				{"privateAudiences": [{"id": "` + planB + `", "type": "subscription"}]}]}`,
			id:   planB,
			want: "in private audience with type",
		},
		{
			name: "no plan has a private audience",
			tree: `{"resources": [{"$schema": "product"}, {"privateAudiences": []}]}`,
			id:   planA,
			want: "not found",
		},
		{
			name: "invalid json response",
			tree: `{"resources": [`,
			id:   planA,
			want: "json decode",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instances[graphResourceIndex].httpClient = &http.Client{
				Transport: &MockHTTPRoundTripper{Transport: map[string]*http.Response{
					treeURL: {StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(tt.tree))},
				}},
			}
			got := Search(tt.id, product)
			if !strings.Contains(got, tt.want) || strings.Contains(got, "marketplace returned") {
				t.Errorf("Search() = %q, want it to report %q", got, tt.want)
			}
		})
	}
}
