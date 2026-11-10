package azure

import (
	"flag"
	"net/http"
	"strings"
	"testing"
)

type operation interface {
	Execute() string
	GetFlagSet() *flag.FlagSet
}

func getCmd() operation       { return NewCommandModule().CommandGetConfig }
func configureCmd() operation { return NewCommandModule().CommandConfigureConfig }
func searchCmd() operation    { return NewCommandModule().CommandSearchConfig }

// countingRoundTripper fails every request and counts how many were attempted.
type countingRoundTripper struct{ calls int }

func (c *countingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	c.calls++
	return nil, http.ErrHandlerTimeout
}

func Test_inputValidation(t *testing.T) {
	prepareConfig()
	prepareTestEnvironment()
	const guid = "12345678-1234-1234-1234-123456789abc"
	tests := []struct {
		name string
		cmd  func() operation
		args []string
		want string
	}{
		{"job id path traversal", getCmd, []string{"-s", "../../../v1.0/users?$top=1#"}, "job id must be a GUID"},
		{"job id not a guid", getCmd, []string{"-s", "job-1"}, "job id must be a GUID"},
		{"subscription id path traversal", getCmd, []string{"-t", "x/resourceGroups"}, "subscription id must be a GUID"},
		{"configure without id", configureCmd, []string{"-o", "offer-2"}, "provide one of -sid or -tid"},
		{"configure with empty sid", configureCmd, []string{"-sid", "", "-o", "offer-2"}, "provide one of -sid or -tid"},
		{"configure with both ids", configureCmd, []string{"-sid", guid, "-tid", guid, "-o", "offer-2"}, "provide only one of -sid or -tid"},
		{"configure non-guid tenant", configureCmd, []string{"-tid", "tenant-1", "-o", "offer-2"}, "tenant id must be a GUID"},
		{"configure non-guid subscription", configureCmd, []string{"-sid", "sub-1", "-o", "offer-2", "-r"}, "subscription id must be a GUID"},
		{"configure unknown offer", configureCmd, []string{"-tid", guid, "-o", "no-such-offer"}, `unknown offer "no-such-offer"`},
		{"search non-guid id", searchCmd, []string{"-id", "x/../y", "-o", "offer-1"}, "id must be a GUID"},
		{"search unknown offer", searchCmd, []string{"-id", guid, "-o", "no-such-offer"}, `unknown offer "no-such-offer"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := &countingRoundTripper{}
			instances[graphResourceIndex].httpClient = &http.Client{Transport: rt}
			instances[managementResourceIndex].httpClient = &http.Client{Transport: rt}

			cmd := tt.cmd()
			if err := cmd.GetFlagSet().Parse(tt.args); err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got := cmd.Execute(); !strings.Contains(got, tt.want) {
				t.Errorf("Execute() = %q, want it to contain %q", got, tt.want)
			}
			if rt.calls != 0 {
				t.Errorf("made %d HTTP call(s) for invalid input, want 0", rt.calls)
			}
		})
	}
}
