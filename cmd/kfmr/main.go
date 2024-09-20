package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/kubeflow/model-registry/pkg/openapi"
	"github.com/tdabasinskas/go-backstage/v2/backstage"
	"os"
	"strconv"
	"time"
)

const (
	RHDH_STATIC_TOKEN    = "RHDH_STATIC_TOKEN"
	BASE_URI             = "/api/model_registry/v1alpha3"
	CREATE_REG_MODEL_URI = "/registered_models"
	// CREATE_MODEL_VERSION_URI can also be '/model_versions' if you do not need to create ModelVersion in RegisteredModel
	CREATE_MODEL_VERSION_URI = "/registered_models/%s/versions"
	CREATE_MODEL_ART_URI     = "/model_artifacts"
	LIST_REG_MODEL_URI       = "/registered_models"
	LIST_MODEL_VERSION_URI   = "/model_versions"
	LIST_MODEL_ART_URI       = "/model_artifacts"
)

func main() {
	backStageRootURL := "https://redhat-developer-hub-ggmtest.apps.gmontero415.devcluster.openshift.com"
	localKubeFlowMRURL := "http://localhost:8081" + BASE_URI
	for n, arg := range os.Args[1:] {
		switch n {
		case 1:
			_, err := strconv.Atoi(arg)
			if err == nil {
				//modeRegistryPort = arg
			}
		case 2:
			//backStageRootURL = arg
		default:
			break
		}
	}
	restyClientKFMR := resty.New()
	if restyClientKFMR == nil {
		fmt.Fprintln(os.Stderr, "could not get kfmr resty client")
		os.Exit(1)
	}
	t := time.Now().UnixNano()
	tstr := strconv.Itoa(int(t))

	modelRegistry := openapi.RegisteredModel{}
	modelRegistry.Name = tstr
	desc := "Description for " + tstr
	modelRegistry.Description = &desc

	registeredModelID := postToModelRegistry(localKubeFlowMRURL+CREATE_REG_MODEL_URI, marshalBody(modelRegistry), restyClientKFMR)

	createURI := fmt.Sprintf(CREATE_MODEL_VERSION_URI, registeredModelID)
	modelVersion := openapi.NewModelVersion(tstr, registeredModelID)
	modelVersion.Description = &desc

	postToModelRegistry(localKubeFlowMRURL+createURI, marshalBody(modelVersion), restyClientKFMR)

	createURI = CREATE_MODEL_ART_URI
	// FYI, ModelArtifact is not sympatico with latest REST API, as the artifactType field is not recognized in the POST,
	// though the response object includes it ... feels like it is hard coded to 'model-artifact'
	//modelArtifact := openapi.ModelArtifact{}
	//modelArtifact.Name = &tstr
	//modelArtifact.Description = &desc
	body := fmt.Sprintf("{\"description\":\"Description for %s\",\"name\":\"%s\"}", tstr, tstr)

	postToModelRegistry(localKubeFlowMRURL+createURI, body, restyClientKFMR)

	fmt.Fprintf(os.Stdout, "registered models %s\n\n", getFromModelRegistry(localKubeFlowMRURL+LIST_REG_MODEL_URI, restyClientKFMR))

	fmt.Fprintf(os.Stdout, "model versions %s\n\n", getFromModelRegistry(localKubeFlowMRURL+LIST_MODEL_VERSION_URI, restyClientKFMR))

	fmt.Fprintf(os.Stdout, "model artifacts %s\n\n", getFromModelRegistry(localKubeFlowMRURL+LIST_MODEL_ART_URI, restyClientKFMR))

	rhdhOAuthGithubToken := os.Getenv(RHDH_STATIC_TOKEN)
	restyClientRHDH := resty.New()
	if restyClientRHDH == nil {
		fmt.Fprintln(os.Stderr, "could not get rhdh resty client")
		os.Exit(1)
	}
	restyClientRHDH.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	resp, err := restyClientRHDH.R().
		SetAuthToken(rhdhOAuthGithubToken).SetHeader("Accept", "application/json").
		Get(backStageRootURL + "/api/catalog/locations")
	if err == nil && resp.StatusCode() == 200 {
		jb, _ := json.MarshalIndent(resp.String(), "", "    ")
		if jb != nil {
			fmt.Fprintf(os.Stdout, "entities %s\n", string(jb))
		} else {
			fmt.Fprintf(os.Stdout, "entities %s\n", resp.String())
		}
		locations := []backstage.LocationListResponse{}
		json.Unmarshal(resp.Body(), &locations)
		if len(locations) == 0 {
			resp, err = restyClientRHDH.R().
				SetAuthToken(rhdhOAuthGithubToken).
				SetBody(map[string]interface{}{"target": "https://github.com/gabemontero/model-catalog/blob/owner-gabemontero/ai-catalog.yaml", "type": "url"}).
				SetHeader("Accept", "application/json").Post(backStageRootURL + "/api/catalog/locations")
			if err != nil {
				fmt.Fprintf(os.Stderr, "get error: %s", err.Error())
				os.Exit(1)
			}
			fmt.Fprintf(os.Stdout, "backstage import produced rc %d and resp string %s\n", resp.StatusCode(), resp.String())

		}
	}

}

func marshalBody(v any) string {
	jb, err := json.Marshal(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "json marshal err %s", err.Error())
		return ""
	}
	return string(jb)
}

func postToModelRegistry(url, body string, client *resty.Client) string {
	resp, _ := client.R().SetBody(body).Post(url)
	postResp := resp.String()
	rc := resp.StatusCode()
	if rc != 200 && rc != 201 {
		fmt.Fprintf(os.Stderr, "%s post with body %s status code %d resp: %s\n", url, body, rc, postResp)
	} else {
		fmt.Fprintf(os.Stdout, "%s post with body %s status code %d resp: %s\n", url, body, rc, postResp)
	}

	retJSON := make(map[string]any)
	err := json.Unmarshal(resp.Body(), &retJSON)
	if err != nil {
		fmt.Fprintf(os.Stderr, "json unmarshall error for %s: %s\n", resp.Body(), err.Error())
	}
	id, ok := retJSON["id"]
	if !ok {
		fmt.Fprintf(os.Stderr, "id fetch did not work for %#v\n", retJSON)
	} else {
		fmt.Fprintf(os.Stdout, "id %s\n", id)
	}
	return fmt.Sprintf("%s", id)
}

func getFromModelRegistry(url string, client *resty.Client) string {
	resp, _ := client.R().Get(url)
	rc := resp.StatusCode()
	getResp := resp.String()
	if rc != 200 {
		fmt.Fprintf(os.Stderr, "get for %s rc %d body %s\n", url, rc, getResp)
	} else {
		fmt.Fprintf(os.Stdout, "get for %s returned ok\n", url)
	}
	jb, err := json.MarshalIndent(getResp, "", "    ")
	if err != nil {
		fmt.Fprint(os.Stderr, "marshall indent error for %s: %s", getResp, err.Error())
	}
	return string(jb)

}
