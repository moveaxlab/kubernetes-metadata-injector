package mutation

import (
	"encoding/json"

	svcmutation "github.com/moveaxlab/kubernetes-metadata-injector/pkg/mutation/service"

	"github.com/sirupsen/logrus"
	"github.com/wI2L/jsondiff"
	corev1 "k8s.io/api/core/v1"
)

// Mutator is a container for mutation
type Mutator struct {
	Logger *logrus.Entry
}

// NewMutator returns an initialised instance of Mutator
func NewMutator(logger *logrus.Entry) *Mutator {
	return &Mutator{Logger: logger}
}

// MutateServiceatch returns a json patch containing all the mutations needed for
// a given service
func (m *Mutator) MutateServicePatch(service *corev1.Service) ([]byte, error) {
	var serviceName string
	if service.Name != "" {
		serviceName = service.Name
	} else {
		if service.ObjectMeta.GenerateName != "" {
			serviceName = service.ObjectMeta.GenerateName
		}
	}
	log := logrus.WithField("service_name", serviceName)

	// list of all mutations to be applied to the service
	mutations := []svcmutation.ServiceMutator{
		svcmutation.InjectAnnotations{Logger: log},
	}

	mservice := service.DeepCopy()

	// apply all mutations
	for _, m := range mutations {
		var err error
		mservice, err = m.Mutate(mservice)
		if err != nil {
			return nil, err
		}
	}

	// generate json patch
	patch, err := jsondiff.Compare(service, mservice)
	if err != nil {
		return nil, err
	}

	patchb, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	return patchb, nil
}
