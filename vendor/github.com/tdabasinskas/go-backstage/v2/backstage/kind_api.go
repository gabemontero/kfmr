package backstage

import (
	"context"
	"net/http"
)

// KindAPI defines name for API kind.
const KindAPI = "API"

// ApiEntityV1alpha1 describes an interface that can be exposed by a component. The API can be defined in different formats,
// like OpenAPI, AsyncAPI, GraphQL, gRPC, or other formats.
// https://github.com/backstage/backstage/blob/master/packages/catalog-model/src/schema/kinds/API.v1alpha1.schema.json
type ApiEntityV1alpha1 struct {
	Entity

	// ApiVersion is always "backstage.io/v1alpha1".
	ApiVersion string `json:"apiVersion" yaml:"apiVersion"`

	// Kind is always "API".
	Kind string `json:"kind" yaml:"kind"`

	// Spec is the specification data describing the API itself.
	Spec *ApiEntityV1alpha1Spec `json:"spec" yaml:"spec"`
}

// ApiEntityV1alpha1Spec describes the specification data describing the API itself.
type ApiEntityV1alpha1Spec struct {
	// Type of the API definition.
	Type string `json:"type" yaml:"type"`

	// Lifecycle state of the API.
	Lifecycle string `json:"lifecycle" yaml:"lifecycle"`

	// Owner is entity reference to the owner of the API.
	Owner string `json:"owner" yaml:"owner"`

	// Definition of the API, based on the format defined by the type.
	Definition string `json:"definition" yaml:"definition"`

	// System is entity reference to the system that the API belongs to.
	System string `json:"system,omitempty" yaml:"system,omitempty"`
}

// apiService handles communication with the API related methods of the Backstage Catalog API.
type apiService typedEntityService[ComponentEntityV1alpha1]

// newApiService returns a new instance of API-type entityService.
func newApiService(s *entityService) *apiService {
	return &apiService{
		client:  s.client,
		apiPath: s.apiPath,
	}
}

// Get returns an API entity identified by the name and the namespace ("default", if not specified) it belongs to.
func (s *apiService) Get(ctx context.Context, n string, ns string) (*ApiEntityV1alpha1, *http.Response, error) {
	cs := (typedEntityService[ApiEntityV1alpha1])(*s)
	return cs.get(ctx, KindAPI, n, ns)
}
