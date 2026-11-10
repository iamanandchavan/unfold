package azure

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/aryannr97/unfold/pkg/helpers"
)

// MSEnableAccountsRes represents the enable accounts response
type MSEnableAccountsRes struct {
	JobID     string        `json:"jobId"`
	JobStatus string        `json:"jobStatus"`
	JobResult string        `json:"jobResult"`
	Errors    []MSErrorResp `json:"errors,omitempty"`
}

// MSErrorResp represents the error response
type MSErrorResp struct {
	Code    interface{}              `json:"code"`
	Message string                   `json:"message"`
	Details []map[string]interface{} `json:"details"`
}

// GetAzureJobStatus calls the MS service to get the status of the Job
func GetAzureJobStatus(jobID string) string {
	if !helpers.IsGUID(jobID) {
		return "job id must be a GUID"
	}
	reqURL := fmt.Sprintf("/rp/product-ingestion/configure/%s/status?$version=2022-07-01", jobID)
	url := instances[graphResourceIndex].BaseURL + reqURL

	resp, err := instances[graphResourceIndex].httpClient.Get(url)
	if err != nil {
		return err.Error()
	}

	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		var res MSEnableAccountsRes
		err := json.NewDecoder(resp.Body).Decode(&res)
		if err != nil {
			return fmt.Sprintf("json decode %v", err.Error())
		}
		errB, _ := json.MarshalIndent(res.Errors, "", " ")
		if res.JobResult == "failed" {
			return fmt.Sprintf("job status error object %v", string(errB))
		}
		resMB, _ := json.MarshalIndent(res, "", " ")
		return fmt.Sprintf("job status response \n%s", string(resMB))
	default:
		b, _ := io.ReadAll(resp.Body)
		errRes := map[string]any{}
		_ = json.Unmarshal(b, &errRes)
		mb, _ := json.MarshalIndent(errRes, "", " ")
		return fmt.Sprintf("marketplace returned %v with response \n%v", resp.StatusCode, string(mb))
	}
}
