package google

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"testing"

	"golang.org/x/net/context"
	cloudidentity "google.golang.org/api/cloudidentity/v1"
	"google.golang.org/api/option"
)

// requestRecorder answers by "METHOD path" and records every request in order.
type requestRecorder struct {
	responses map[string]string
	requests  []string
}

func (r *requestRecorder) RoundTrip(req *http.Request) (*http.Response, error) {
	key := req.Method + " " + req.URL.Path
	r.requests = append(r.requests, key)
	if body, ok := r.responses[key]; ok {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(body))}, nil
	}
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(bytes.NewBufferString(`{"error": {"code": 404, "message": "not found", "status": "NOT_FOUND"}}`)),
	}, nil
}

func useRecorder(responses map[string]string) *requestRecorder {
	rec := &requestRecorder{responses: responses}
	instance.CloudIdentityService, _ = cloudidentity.NewService(context.Background(),
		option.WithHTTPClient(&http.Client{Transport: rec}))
	instance.Groups = make(map[string]*cloudidentity.LookupGroupNameResponse)
	return rec
}

const (
	groupLookup      = "GET /v1/groups:lookup"
	membershipLookup = "GET /v1/groups/test-group/memberships:lookup"
	membershipGet    = "GET /v1/groups/test-group/memberships/xyz"
	membershipDelete = "DELETE /v1/groups/test-group/memberships/xyz"
)

func Test_membershipUsesLookupInsteadOfListing(t *testing.T) {
	Config.clientOpts = nil
	prepareTestEnvironment()
	responses := map[string]string{
		groupLookup:      `{"name": "groups/test-group"}`,
		membershipLookup: `{"name": "groups/test-group/memberships/xyz"}`,
		membershipGet:    `{"name": "groups/test-group/memberships/xyz", "preferredMemberKey": {"id": "test-id"}, "roles": [{"name": "MEMBER"}]}`,
		membershipDelete: `{"done": true}`,
	}

	t.Run("search", func(t *testing.T) {
		rec := useRecorder(responses)
		m, err := CheckGroupMembershipForEmailIDs("test-group", "test-id")
		if err != nil || m.Name != "groups/test-group/memberships/xyz" || len(m.Roles) != 1 {
			t.Fatalf("got membership %+v, err %v", m, err)
		}
		if want := []string{groupLookup, membershipLookup, membershipGet}; !reflect.DeepEqual(rec.requests, want) {
			t.Errorf("requests = %v, want %v", rec.requests, want)
		}
	})

	t.Run("remove", func(t *testing.T) {
		rec := useRecorder(responses)
		if err := RemoveMemberFromGroupID("test-group", "test-id"); err != nil {
			t.Fatal(err)
		}
		if want := []string{groupLookup, membershipLookup, membershipDelete}; !reflect.DeepEqual(rec.requests, want) {
			t.Errorf("requests = %v, want %v", rec.requests, want)
		}
	})

	t.Run("lookup failure other than not found", func(t *testing.T) {
		useRecorder(map[string]string{groupLookup: responses[groupLookup]})
		instance.CloudIdentityService, _ = cloudidentity.NewService(context.Background(),
			option.WithHTTPClient(&http.Client{Transport: &MockHTTPRoundTripper{
				Transport: map[string]*http.Response{
					"/v1/groups:lookup": {StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(responses[groupLookup]))},
					"/v1/groups/test-group/memberships:lookup": {
						StatusCode: http.StatusForbidden,
						Body:       io.NopCloser(bytes.NewBufferString(`{"error": {"code": 403, "message": "permission denied"}}`)),
					},
				},
			}}))
		_, err := CheckGroupMembershipForEmailIDs("test-group", "test-id")
		if err == nil || err.Error() == "member not found" {
			t.Errorf("err = %v, want the permission error", err)
		}
	})

	t.Run("not a member", func(t *testing.T) {
		useRecorder(map[string]string{groupLookup: responses[groupLookup]})
		if _, err := CheckGroupMembershipForEmailIDs("test-group", "someone-else"); err == nil || err.Error() != "member not found" {
			t.Errorf("err = %v, want member not found", err)
		}
	})
}
