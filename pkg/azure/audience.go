package azure

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/aryannr97/unfold/pkg/helpers"
)

// Resources represents the resources returned by the Partner Center API
type Resources struct {
	Resources []TreeResource `json:"resources"`
}

// TreeResource represents the tree resource returned by the Partner Center API
type TreeResource struct {
	PrivateAudiences []PrivateAudience `json:"privateAudiences"`
}

// PrivateAudience represents the private audience returned by the Partner Center API
type PrivateAudience struct {
	Audtype string `json:"type"`
	ID      string `json:"id"`
}

// Search searches for a given id in the private audience list for a specified offer
func Search(id string, offer string) string {
	if offer == "" {
		return "offer cannot be empty"
	}
	resource, err := GetPrivateAudienceListForOffer(offer)
	if err != nil {
		return err.Error()
	}

	for _, obj := range resource.PrivateAudiences {
		if obj.ID == id {
			return fmt.Sprintf("found %s in private audience with type %s", helpers.GreenValue(id), helpers.GreenValue(obj.Audtype))
		}
	}
	return fmt.Sprintf("given id %s in private audience", helpers.RedValue("not found"))
}

// GetPrivateAudienceListForOffer makes a GET request to the Partner Center API
// to retrieve the private audience entries of all plans of a specified offer.
func GetPrivateAudienceListForOffer(offerID string) (TreeResource, error) {
	reqURL := fmt.Sprintf("/rp/product-ingestion/resource-tree/product/%s", offerID)
	url := instances[graphResourceIndex].BaseURL + reqURL

	resp, httpErr := instances[graphResourceIndex].httpClient.Get(url)
	if httpErr != nil {
		return TreeResource{}, httpErr
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return TreeResource{}, fmt.Errorf("marketplace returned %d with response \n%v", resp.StatusCode, helpers.GetErrorResponseBody(b))
	}

	var res Resources
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return TreeResource{}, fmt.Errorf("json decode %w", err)
	}

	// Private audiences are configured per plan, so merge every plan's
	// audience; an offer without any audience is a valid, empty result.
	var merged TreeResource
	for _, obj := range res.Resources {
		merged.PrivateAudiences = append(merged.PrivateAudiences, obj.PrivateAudiences...)
	}
	return merged, nil
}
