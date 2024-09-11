package main

import (
     "context"
     "fmt"
     "github.com/kubeflow/model-registry/pkg/api"
     "github.com/kubeflow/model-registry/pkg/core"
     "github.com/kubeflow/model-registry/pkg/openapi"
     "google.golang.org/grpc"
     "google.golang.org/grpc/credentials/insecure"
     "log"
     "os"
     "strconv"
)

func main() {
     port := "9090"
     for _, arg := range os.Args[1:] {
          _, err := strconv.Atoi(arg)
          if err == nil {
               port = arg
               break
          }
     }
     err := testGRPCToKubeFlowModelRegistry(port)
     if err != nil {
          os.Exit(1)
     }
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
