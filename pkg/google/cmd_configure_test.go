package google

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/net/context"
	cloudidentity "google.golang.org/api/cloudidentity/v1"
	"google.golang.org/api/option"
)

func Test_commandConfigureConfig_Execute(t *testing.T) {
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
			name: "test configure command add member success",
			args: []string{"-g", "test-group", "-id", "test-id"},
			transport: map[string]*http.Response{
				"/v1/groups:lookup": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"name": "groups/test-group"}`)),
				},
				"/v1/groups/test-group/memberships": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"name": "success"}`)),
				},
			},
			want: "successfully added the member to the given group",
		},
		{
			name:          "test configure command group not found",
			args:          []string{"-g", "test-group", "-id", "test-id"},
			transport:     nil,
			httpCallError: errors.New("http call error"),
			want:          "failed to add member to the given group",
		},
		{
			name: "test configure command add member error",
			args: []string{"-g", "test-group", "-id", "test-id"},
			transport: map[string]*http.Response{
				"/v1/groups:lookup": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"name": "groups/test-group"}`)),
				},
			},
			httpCallError: errors.New("http call error"),
			errorOnIndex:  1,
			want:          "failed to add member to the given group",
		},
		{
			name: "test configure command remove member success",
			args: []string{"-g", "test-group", "-id", "test-id", "-r"},
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
					Body:       io.NopCloser(bytes.NewBufferString(`{"name": "success"}`)),
				},
			},
			want: "successfully removed the member from the group",
		},
		{
			name: "test configure command membership not found",
			args: []string{"-g", "test-group", "-id", "test-id", "-r"},
			transport: map[string]*http.Response{
				"/v1/groups:lookup": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"name": "groups/test-group"}`)),
				},
			},
			httpCallError: errors.New("http call error"),
			errorOnIndex:  1,
			want:          "unable to remove the member",
		},
		{
			name: "test configure command remove member error",
			args: []string{"-g", "test-group", "-id", "test-id", "-r"},
			transport: map[string]*http.Response{
				"/v1/groups:lookup": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"name": "groups/test-group"}`)),
				},
				"/v1/groups/test-group/memberships:lookup": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"name": "groups/test-group/memberships/xyz"}`)),
				},
			},
			httpCallError: errors.New("http call error"),
			errorOnIndex:  2,
			want:          "unable to remove the member",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCommandModule().CommandConfigureConfig
			c.GetFlagSet().Parse(tt.args)
			instance.CloudIdentityService, _ = cloudidentity.NewService(context.Background(),
				option.WithHTTPClient(&http.Client{
					Transport: &MockHTTPRoundTripper{
						Transport:    tt.transport,
						Error:        tt.httpCallError,
						ErrorOnIndex: tt.errorOnIndex,
					},
				}),
			)
			if got := c.Execute(); !strings.Contains(got, tt.want) {
				t.Errorf("commandConfigureConfig.Execute() = %v, want %v", got, tt.want)
			}
			instance.Groups = make(map[string]*cloudidentity.LookupGroupNameResponse)
		})
	}
}
