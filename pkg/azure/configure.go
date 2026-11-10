package azure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aryannr97/unfold/pkg/helpers"
)

// MSGraphEnableAccount represents the enable account request body
type MSGraphEnableAccount struct {
	Schema    string       `json:"$schema"`
	Resources []MSResource `json:"resources"`
}

// MSResource represents the resource request body
type MSResource struct {
	Schema           string            `json:"$schema"`
	Product          string            `json:"product"`
	Plan             string            `json:"plan"`
	PrivateAudiences MSPrivateAudience `json:"privateAudiences"`
}

// MSPrivateAudience represents the private audience request body
type MSPrivateAudience struct {
	Add    []MSProperty `json:"add"`
	Remove []MSProperty `json:"remove"`
}

// MSProperty represents the property request body
type MSProperty struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// LoggerObj summarizes the response in a structured format
type LoggerObj struct {
	SubscriptionID   string `json:"subscriptionID"`
	TenantID         string `json:"tenantID,omitempty"`
	SyncAudienceType string `json:"syncAudienceType"`
	AzureJobID       string `json:"azureJobID"`
	AzureJobResult   string `json:"azureJobResult"`
}

// MakeConfigurationRequest decides type of audience to be used for syncing and make request to Azure
func MakeConfigurationRequest(image, id, audType, mode string) string {
	loggerObj := LoggerObj{}

	audienceList := []MSProperty{}
	if strings.EqualFold(audType, "tenant") {
		loggerObj.TenantID = id
		loggerObj.SyncAudienceType = "tenant"
		audienceList = append(audienceList, MSProperty{
			Type: "tenant",
			ID:   id,
		})
	} else {
		loggerObj.SubscriptionID = id
		loggerObj.SyncAudienceType = "subscription"
		audienceList = append(audienceList, MSProperty{
			Type: "subscription",
			ID:   id,
		})
	}

	// fetch all plans for offer/image
	plans, httpErr := getPlans(config.Offers[image].ProductDurableID)
	if httpErr != nil {
		return httpErr.Error()
	}

	reqBody := prepareRequestBody(image, plans, audienceList, mode)

	// make request to Azure
	azureJob, err := configurePrivateAudienceAPI(reqBody)
	if err != nil {
		return err.Error()
	}

	loggerObj.AzureJobID = azureJob.JobID
	loggerObj.AzureJobResult = azureJob.JobResult

	b, _ := json.MarshalIndent(loggerObj, "", " ")

	return fmt.Sprintf("configure response \n%v", string(b))
}

// prepareRequestBody returns requestBody to be used for syncing private audience
func prepareRequestBody(image string, plans []string, audienceList []MSProperty, mode string) MSGraphEnableAccount {
	body := MSGraphEnableAccount{
		Schema:    "https://schema.mp.microsoft.com/schema/configure/2022-03-01-preview2",
		Resources: make([]MSResource, 0, len(plans)),
	}

	// Only one side of the private audience change may carry the ids;
	// sending them in both add and remove makes the request ambiguous.
	audiences := MSPrivateAudience{Add: []MSProperty{}, Remove: []MSProperty{}}
	if mode == RemoveMode {
		audiences.Remove = audienceList
	} else {
		audiences.Add = audienceList
	}

	product := "product/" + config.Offers[image].ProductDurableID
	for _, planID := range plans {
		body.Resources = append(body.Resources, MSResource{
			Schema:           "https://schema.mp.microsoft.com/schema/price-and-availability-update-private-audiences/2022-03-01-preview2",
			Product:          product,
			Plan:             planID,
			PrivateAudiences: audiences,
		})
	}

	return body
}

// configurePrivateAudienceAPI makes an actual API call to Azure for syncing private audience
func configurePrivateAudienceAPI(reqBody MSGraphEnableAccount) (*MSEnableAccountsRes, error) {
	b, _ := json.Marshal(reqBody)

	reqURL := "/rp/product-ingestion/configure?$version=2022-03-01-preview2"

	body := bytes.NewBuffer(b)

	url := instances[graphResourceIndex].BaseURL + reqURL

	resp, err := instances[graphResourceIndex].httpClient.Post(url, "application/json", body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var res MSEnableAccountsRes
		err := json.NewDecoder(resp.Body).Decode(&res)
		if err != nil {
			return nil, fmt.Errorf("json decode %s", err.Error())
		}
		return &res, nil
	default:
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("marketplace returned %v for configurePrivateAudienceAPI with response %v", resp.StatusCode, helpers.GetErrorResponseBody(b))
	}
}

type Plans struct {
	Value []Resource `json:"value"`
}

// Resource is a singular Azure resource type entity.
// Each type is described using a dedicated schema definition as referenced by the "$schema" property.
// Other configuaration properties are not covered here as we only require unique identification with "id" field.
type Resource struct {
	Schema string `json:"$schema"`
	ID     string `json:"id"`
}

// getPlans return unique planIDs associated with product durable id of an offer/image.
func getPlans(productID string) ([]string, error) {
	reqURL := fmt.Sprintf("/rp/product-ingestion/plan?product=product/%s&$version=2022-03-01-preview2", productID)
	url := instances[graphResourceIndex].BaseURL + reqURL

	resp, err := instances[graphResourceIndex].httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var res Plans
		var ids []string
		err := json.NewDecoder(resp.Body).Decode(&res)
		if err != nil {
			return nil, fmt.Errorf("json decode %s", err.Error())
		}

		for _, plan := range res.Value {
			ids = append(ids, plan.ID)
		}

		return ids, nil
	default:
		b, _ := json.Marshal(resp.Body)
		return nil, fmt.Errorf("marketplace returned %v for getPlans with response %v", resp.StatusCode, string(b))
	}
}
