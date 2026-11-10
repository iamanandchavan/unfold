package google

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/net/context"
	cloudidentity "google.golang.org/api/cloudidentity/v1"
	"google.golang.org/api/option"

	"github.com/aryannr97/unfold/pkg/helpers"
)

func Test_searchMembershipRoles(t *testing.T) {
	Config.clientOpts = nil // config_test.go leaves incompatible options behind
	prepareTestEnvironment()
	tests := []struct {
		name, roles, want string
	}{
		{"membership without roles", `[]`, helpers.GreenValue("NONE")},
		{"membership with several roles", `[{"name": "MEMBER"}, {"name": "OWNER"}]`, helpers.GreenValue("MEMBER,OWNER")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := map[string]*http.Response{
				"/v1/groups:lookup": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"name": "groups/test-group"}`)),
				},
				"/v1/groups/test-group/memberships:lookup": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"name": "groups/test-group/memberships/xyz"}`)),
				},
				"/v1/groups/test-group/memberships/xyz": {
					StatusCode: http.StatusOK,
					Body: io.NopCloser(bytes.NewBufferString(`{"name": "groups/test-group/memberships/xyz",
						"preferredMemberKey": {"id": "test-id"}, "roles": ` + tt.roles + `}`)),
				},
			}
			instance.CloudIdentityService, _ = cloudidentity.NewService(context.Background(),
				option.WithHTTPClient(&http.Client{Transport: &MockHTTPRoundTripper{Transport: transport}}))
			defer func() { instance.Groups = make(map[string]*cloudidentity.LookupGroupNameResponse) }()

			c := NewCommandModule().CommandSearchConfig
			if err := c.GetFlagSet().Parse([]string{"-g", "test-group", "-id", "test-id"}); err != nil {
				t.Fatal(err)
			}
			if got := c.Execute(); !strings.Contains(got, "emailID is found to be "+tt.want+" of the group") {
				t.Errorf("Execute() = %q, want role %q", got, tt.want)
			}
		})
	}
}
