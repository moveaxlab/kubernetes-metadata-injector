// Package admission handles kubernetes admissions,
// it takes admission requests and returns admission reviews;
package admission

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/moveaxlab/kubernetes-metadata-injector/pkg/mutation"
	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

// Admitter is a container for admission business
type Admitter struct {
	Logger  *logrus.Entry
	Request *admissionv1.AdmissionRequest
}

// MutateServiceReview takes an admission request and mutates the service within,
// it returns an admission review with mutations as a json patch (if any)
func (a Admitter) MutateServiceReview() (*admissionv1.AdmissionReview, error) {
	service, err := a.Service()
	if err != nil {
		e := fmt.Sprintf("could not parse service in admission review request: %v", err)
		return reviewResponse(a.Request.UID, false, http.StatusBadRequest, e), err
	}

	m := mutation.NewMutator(a.Logger)
	patch, err := m.MutateServicePatch(service)
	if err != nil {
		e := fmt.Sprintf("could not mutate service: %v", err)
		return reviewResponse(a.Request.UID, false, http.StatusBadRequest, e), err
	}

	return patchReviewResponse(a.Request.UID, patch)
}

// Service extracts a service from an admission request
func (a Admitter) Service() (*corev1.Service, error) {
	if a.Request.Kind.Kind != "Service" {
		return nil, fmt.Errorf("only services are supported here")
	}

	s := corev1.Service{}
	if err := json.Unmarshal(a.Request.Object.Raw, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

// reviewResponse TODO: godoc
func reviewResponse(uid types.UID, allowed bool, httpCode int32,
	reason string) *admissionv1.AdmissionReview {
	return &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     uid,
			Allowed: allowed,
			Result: &metav1.Status{
				Code:    httpCode,
				Message: reason,
			},
		},
	}
}

// patchReviewResponse builds an admission review with given json patch
func patchReviewResponse(uid types.UID, patch []byte) (*admissionv1.AdmissionReview, error) {
	patchType := admissionv1.PatchTypeJSONPatch

	return &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:       uid,
			Allowed:   true,
			PatchType: &patchType,
			Patch:     patch,
		},
	}, nil
}
