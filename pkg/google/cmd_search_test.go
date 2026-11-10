package google

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/aryannr97/unfold/pkg/helpers"
	"google.golang.org/api/cloudidentity/v1"
	"google.golang.org/api/option"
)

func Test_commandSearchConfig_Execute(t *testing.T) {
	prepareTestEnvironment()
	tests := []struct {
		name          string
		args          []string
		transport     map[string]*http.Response
		httpCallError error
		errorOnIndex  int
		want          string
	}{
		{
			name: "test search command",
			args: []string{"-g", "test-group", "-id", "test-id"},
			transport: map[string]*http.Response{
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
					Body:       io.NopCloser(bytes.NewBufferString(`{"name": "groups/test-group/memberships/xyz", "preferredMemberKey": {"id": "test-id"}, "roles": [{"name": "MEMBER"}]}`)),
				},
			},
			want: "emailID is found to be " + helpers.GreenValue("MEMBER") + " of the group with membership name " + helpers.GreenValue("groups/test-group/memberships/xyz"),
		},
		{
			name:          "test search command membership call error, with existing group lookup",
			args:          []string{"-g", "test-group", "-id", "test-id"},
			transport:     nil,
			httpCallError: errors.New("http call error"),
			// error on index 0 because the group lookup is successful
			errorOnIndex: 0,
			want:         "http call error",
		},
		{
			name:          "test search command group not found",
			args:          []string{"-g", "test-group", "-id", "test-id"},
			transport:     nil,
			httpCallError: errors.New("http call error"),
			want:          "http call error",
		},
		{
			name: "test search command membership not found",
			args: []string{"-g", "test-group", "-id", "test-id"},
			transport: map[string]*http.Response{
				"/v1/groups:lookup": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"name": "groups/test-group"}`)),
				},
			},
			want: "member not found",
		},
		{
			name:          "test search command flags not provided",
			args:          []string{},
			transport:     nil,
			httpCallError: nil,
			want:          "something went wrong",
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCommandModule().CommandSearchConfig
			c.GetFlagSet().Parse(tt.args)
			instance.CloudIdentityService, _ = cloudidentity.NewService(context.Background(),
				option.WithHTTPClient(&http.Client{
					Transport: &MockHTTPRoundTripper{
						Transport:    tt.transport,
						Error:        tt.httpCallError,
						ErrorOnIndex: tt.errorOnIndex,
					},
				}))
			if got := c.Execute(); !strings.Contains(got, tt.want) {
				t.Errorf("commandSearchConfig.Execute() = %v, want %v", got, tt.want)
			}
			// reset the groups map after each test except for the first test
			// avoid using the transport again and do group lookup instead, stored from previous test
			// done specifically to cover func (s *Service) GetGroup(string) *cloudidentity.LookupGroupNameResponse
			if i > 0 {
				instance.Groups = make(map[string]*cloudidentity.LookupGroupNameResponse)
			}
		})
	}
}
