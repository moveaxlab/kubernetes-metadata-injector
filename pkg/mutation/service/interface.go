package service

import corev1 "k8s.io/api/core/v1"

// serviceMutator is an interface used to group functions mutating services
type ServiceMutator interface {
	Mutate(*corev1.Service) (*corev1.Service, error)
	Name() string
}
