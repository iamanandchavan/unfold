package azure

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// recordingRoundTripper serves canned responses by URL and records request bodies by URL.
type recordingRoundTripper struct {
	responses map[string]string
	bodies    map[string][]byte
}

func (r *recordingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	url := req.URL.String()
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		r.bodies[url] = b
	}
	body, ok := r.responses[url]
	if !ok {
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(bytes.NewBufferString(`{}`))}, nil
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(body))}, nil
}

func Test_MakeConfigurationRequest_body(t *testing.T) {
	prepareConfig()
	prepareTestEnvironment()
	const (
		id         = "12345678-1234-1234-1234-123456789abc"
		plansURL   = "https://graph.microsoft.com/rp/product-ingestion/plan?product=product/87654321-4321-4321-4321-210987654321&$version=2022-03-01-preview2"
		configURL  = "https://graph.microsoft.com/rp/product-ingestion/configure?$version=2022-03-01-preview2"
		plansBody  = `{"value": [{"id": "plan-1"}, {"id": "plan-2"}]}`
		configBody = `{"jobId": "job-1", "jobStatus": "pending", "jobResult": "pending"}`
	)
	tests := []struct {
		name, audType, mode string
		wantAdd, wantRemove []MSProperty
	}{
		{"add tenant", "tenant", AddMode, []MSProperty{{Type: "tenant", ID: id}}, []MSProperty{}},
		{"remove tenant", "tenant", RemoveMode, []MSProperty{}, []MSProperty{{Type: "tenant", ID: id}}},
		{"add subscription", "sub", AddMode, []MSProperty{{Type: "subscription", ID: id}}, []MSProperty{}},
		{"remove subscription", "sub", RemoveMode, []MSProperty{}, []MSProperty{{Type: "subscription", ID: id}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := &recordingRoundTripper{
				responses: map[string]string{plansURL: plansBody, configURL: configBody},
				bodies:    map[string][]byte{},
			}
			instances[graphResourceIndex].httpClient = &http.Client{Transport: rt}

			MakeConfigurationRequest("offer-2", id, tt.audType, tt.mode)

			var got MSGraphEnableAccount
			if err := json.Unmarshal(rt.bodies[configURL], &got); err != nil {
				t.Fatalf("configure request body not captured or invalid: %v", err)
			}
			if len(got.Resources) != 2 {
				t.Fatalf("want one resource per plan (2), got %d", len(got.Resources))
			}
			for _, res := range got.Resources {
				gotAdd, _ := json.Marshal(res.PrivateAudiences.Add)
				wantAdd, _ := json.Marshal(tt.wantAdd)
				gotRemove, _ := json.Marshal(res.PrivateAudiences.Remove)
				wantRemove, _ := json.Marshal(tt.wantRemove)
				if !bytes.Equal(gotAdd, wantAdd) || !bytes.Equal(gotRemove, wantRemove) {
					t.Errorf("plan %s: add=%s remove=%s, want add=%s remove=%s",
						res.Plan, gotAdd, gotRemove, wantAdd, wantRemove)
				}
			}
		})
	}
}
