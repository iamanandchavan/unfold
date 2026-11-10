package azure

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/aryannr97/unfold/pkg/helpers"
)

// tenantFinder uses regex to find tenant ids from azure api response
type TenantFinder struct {
	uniTenantRegex   *regexp.Regexp
	multiTenantRegex *regexp.Regexp
}

// NewTenantFinder returns a new tenant finder
func NewTenantFinder() *TenantFinder {
	return &TenantFinder{}
}

// fillDefaults populates default regular expressions
func (a *TenantFinder) fillDefaults() {
	a.uniTenantRegex = regexp.MustCompile(`tenant \'https:\/\/sts\.windows\.net\/(\S+)\/\' associated`)
	a.multiTenantRegex = regexp.MustCompile(`tenants (\'https:\/\/sts\.windows\.net\/(\S+)\/,?\')+ associated`)
}

// RetrieveTenantIDs uses regex to find tenant ids from azure api response
func (a *TenantFinder) retrieveTenantIDsFromErrMsg(msg string) []string {
	// load predefined regex
	a.fillDefaults()

	tids := []string{}

	// try to find single tenant
	res := a.uniTenantRegex.FindStringSubmatch(msg)

	// if found, return slice with single element
	if len(res) > 0 {
		tids = append(tids, res[1])
		return tids
	}

	// try to find multiple tenants
	res = a.multiTenantRegex.FindStringSubmatch(msg)
	// It is assumed that this case only happens for more than one tenant.
	if len(res) > 1 {
		urls := strings.Split(strings.Trim(res[1], "'"), ",")

		for _, url := range urls {
			tmp := strings.Split(url, "/")
			tids = append(tids, tmp[3])
		}
	}

	return tids
}

// GetTenantBySubscriptionID calls azure API to get tenantID specific to subscription id.
// It works on the recommended logic from Microsoft to apply regex on error response of the API.
func (a *TenantFinder) GetTenantBySubscriptionID(id string) ([]string, error) {
	if !helpers.IsGUID(id) {
		return nil, errors.New("subscription id must be a GUID")
	}
	resBody := ErrResponse{}
	reqURL := fmt.Sprintf("/subscriptions/%s?api-version=2022-12-01", id)

	url := instances[managementResourceIndex].BaseURL + reqURL
	resp, _ := instances[managementResourceIndex].httpClient.Get(url)

	// Precautionary check for server timeouts or outages
	if resp != nil && resp.Body != nil {
		// Decode the response body to the standard error format
		err := json.NewDecoder(resp.Body).Decode(&resBody)
		if err != nil {
			return []string{}, fmt.Errorf("json decode %w", err)
		}
	}

	defer resp.Body.Close()

	return a.retrieveTenantIDsFromErrMsg(resBody.Error.Message), nil
}

// ErrResponse is the standard Azure API error response
type ErrResponse struct {
	Error Error `json:"error"`
}

// Error refers to the error object of Azure API response
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
