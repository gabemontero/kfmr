package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/kubeflow/model-registry/pkg/api"
	"github.com/kubeflow/model-registry/pkg/core"
	"github.com/kubeflow/model-registry/pkg/openapi"
	"github.com/tdabasinskas/go-backstage/v2/backstage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	RHDH_GITHUB_OAUTH_TOKEN = "RHDH_GITHUB_OAUTH_TOKEN"
)

func main() {
	//modeRegistryPort := "9090"
	//backStageRootURL := "https://backstage-developer-hub-ggmtest.apps.gmontero415.devcluster.openshift.com"
	localKubeFlowMRURL := "http://localhost:8081/api/model_registry/v1alpha3"
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
	createURI := "/registered_models"
	createURL := localKubeFlowMRURL + createURI
	createBody := fmt.Sprintf("{ \"name\": \"%s\", \"description\": \"%s description\" }", tstr, tstr)

	resp, _ := restyClientKFMR.R().SetBody(createBody).Post(createURL)
	postResp := resp.String()
	rc := resp.StatusCode()
	if rc != 200 || rc != 201 {
		fmt.Fprintf(os.Stderr, "reg model post status code %d resp: %s", rc, postResp)
	} else {
		fmt.Fprintf(os.Stdout, "reg model post status code %d resp: %s", rc, postResp)
	}

	retJSON := make(map[string]string)
	err := json.Unmarshal(resp.Body(), &retJSON)
	if err != nil {
		fmt.Fprintf(os.Stderr, "json unmarshall error: %s", err.Error())
	}
	registeredModelID, ok := retJSON["id"]
	if !ok {
		fmt.Fprintf(os.Stderr, "reg model fetch for id did not work")
	} else {
		fmt.Fprintf(os.Stdout, "reg model id %s", registeredModelID)
	}

	createURI = fmt.Sprintf("/registered_models/%s/versions", registeredModelID)
	createBody = fmt.Sprintf("{ \"name\": \"%s\", \"description\": \"%s description\", \"registeredModelId\": \"%s\" }", tstr, tstr, registeredModelID)
	createURL = localKubeFlowMRURL + createURI
	resp, _ = restyClientKFMR.R().SetBody(createBody).Post(createURL)
	postResp = resp.String()
	rc = resp.StatusCode()
	if rc != 200 || rc != 201 {
		fmt.Fprintf(os.Stderr, "model version post status code %d resp: %s", rc, postResp)
	} else {
		fmt.Fprintf(os.Stdout, "model version post status code %d resp: %s", rc, postResp)
	}

	createURI = "/model_artifacts"
	createBody = fmt.Sprintf("{ \"name\": \"%s\", \"description\": \"%s description\" }", tstr, tstr)
	createURL = localKubeFlowMRURL + createURI
	resp, _ = restyClientKFMR.R().SetBody(createBody).Post(createURL)
	postResp = resp.String()
	rc = resp.StatusCode()
	if rc != 200 || rc != 201 {
		fmt.Fprintf(os.Stderr, "model artifacts post status code %d resp: %s", rc, postResp)
	} else {
		fmt.Fprintf(os.Stdout, "model artifacts post status code %d resp: %s", rc, postResp)
	}

	resp, _ = restyClientKFMR.R().Get(localKubeFlowMRURL + "/registered_models")
	fmt.Fprintf(os.Stdout, "registered models %s\n\n", resp.String())

	resp, _ = restyClientKFMR.R().Get(localKubeFlowMRURL + "/model_versions")
	fmt.Fprintf(os.Stdout, "model versions %s\n\n", resp.String())

	resp, _ = restyClientKFMR.R().Get(localKubeFlowMRURL + "/model_artifacts")
	fmt.Fprintf(os.Stdout, "model artifacts %s\n\n", resp.String())

	//rhdhOAuthGithubToken := os.Getenv(RHDH_GITHUB_OAUTH_TOKEN)
	//restyClientRHDH := resty.New()
	//if restyClientRHDH == nil {
	//	fmt.Fprintln(os.Stderr, "could not get rhdh resty client")
	//	os.Exit(1)
	//}
	//restyClientRHDH.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	//
	//resp, err = restyClientRHDH.R().
	//	SetAuthToken(rhdhOAuthGithubToken).
	//	SetHeader("Accept", "application/json").Get(backStageRootURL + "/api/catalog/entities")
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, "get error: %s", err.Error())
	//	os.Exit(1)
	//}

	component := backstage.ComponentEntityV1alpha1{}
	fmt.Fprintf(os.Stdout, "component %#v\n\n", component)

	//err = testGRPCToKubeFlowModelRegistry(modeRegistryPort)
	//if err != nil {
	//	os.Exit(1)
	//}
}

func testGRPCToKubeFlowModelRegistry(port string) error {
	conn, err := grpc.DialContext(
		context.Background(),
		fmt.Sprintf("localhost:%s", port),
		grpc.WithReturnConnectionError(),
		grpc.WithBlock(), // optional
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("error dialing connection to mlmd server localhost:9090: %v", err)
	}
	defer conn.Close()

	service, err := core.NewModelRegistryService(conn)
	if err != nil {
		fmt.Fprintf(os.Stdout, "error creating model registry core service: %v", err)
	}

	modelName := "MODEL_NAME"
	modelDescription := "MODEL_DESCRIPTION"

	// for now the name must match (i.e. no regexp or wildcards
	registeredModel, err := service.GetRegisteredModelByParams(&modelName, nil)
	if err != nil {
		log.Printf("unable to find model %s: %v", modelName, err)
		// register a new model
		registeredModel, err = service.UpsertRegisteredModel(&openapi.RegisteredModel{
			Name:        modelName,
			Description: &modelDescription,
		})
		if err != nil {
			fmt.Fprintf(os.Stdout, "error registering model: %v", err)
			return err
		}
		// register model version
		versionName := "VERSION_NAME"
		versionDescription := "VERSION_DESCRIPTION"
		versionScore := 0.83

		modelVersion, err := service.UpsertModelVersion(&openapi.ModelVersion{
			Name:        versionName,
			Description: &versionDescription,
			CustomProperties: &map[string]openapi.MetadataValue{
				"score": {
					MetadataDoubleValue: &openapi.MetadataDoubleValue{
						DoubleValue: versionScore,
					},
				},
			},
		}, registeredModel.Id)
		if err != nil {
			fmt.Fprintf(os.Stdout, "error registering model version: %v", err)
			return err
		}

		artifactName := "ARTIFACT_NAME"
		artifactDescription := "ARTIFACT_DESCRIPTION"
		artifactUri := "ARTIFACT_URI"

		// register model artifact
		modelArtifact, err := service.UpsertModelArtifact(&openapi.ModelArtifact{
			Name:        &artifactName,
			Description: &artifactDescription,
			Uri:         &artifactUri,
		}, modelVersion.Id)
		if err != nil {
			fmt.Fprintf(os.Stdout, "error creating model artifact: %v", err)
		} else {
			m, e := modelArtifact.ToMap()
			if e != nil {
				fmt.Println(fmt.Sprintf("error: %s", e.Error()))
			}
			fmt.Println(fmt.Sprintf("%#v", m))
		}
	}

	allVersions, err := service.GetModelVersions(api.ListOptions{}, registeredModel.Id)
	if err != nil {
		fmt.Fprintf(os.Stdout, "error retrieving model versions for model %s: %v", *registeredModel.Id, err)
	}
	if allVersions != nil {
		for _, v := range allVersions.GetItems() {
			m, e := v.ToMap()
			if e != nil {
				fmt.Println("error: %s", e.Error())
			}
			fmt.Println(fmt.Sprintf("%#v", m))
		}
	}
	return nil
}
