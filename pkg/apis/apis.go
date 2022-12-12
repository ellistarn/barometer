package apis

import (
	_ "embed"

	"github.com/ellistarn/barometer/pkg/apis/v1alpha1"
	"github.com/samber/lo"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	// Builder includes all types within the apis package
	Builder = runtime.NewSchemeBuilder(
		v1alpha1.SchemeBuilder.AddToScheme,
	)
	// AddToScheme may be used to add all resources defined in the project to a Scheme
	AddToScheme = Builder.AddToScheme
)

//go:generate controller-gen crd object:headerFile="../../hack/boilerplate.go.txt" paths="./..." output:crd:artifacts:config=crds
var (
	//go:embed crds/barometer.io_barometers.yaml
	BarometerCRD []byte
	CRDs           = []*v1.CustomResourceDefinition{
		lo.Must(Unmarshal[v1.CustomResourceDefinition](BarometerCRD)),
	}
)

func Unmarshal[T any](raw []byte) (*T, error) {
	t := *new(T)
	if err := yaml.Unmarshal(raw, &t); err != nil {
		return nil, err
	}
	return &t, nil
}
