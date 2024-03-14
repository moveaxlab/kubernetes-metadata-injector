package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestInjectAnnotationsMutateNoAnnotations(t *testing.T) {
	want := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name: "test",
			Annotations: map[string]string{
				"test": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       "https",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			}},
		},
	}

	svc := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name: "test",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       "https",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			}},
		},
	}

	got, err := InjectAnnotations{Logger: logger()}.Mutate(svc)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, want, got)
}

func TestInjectAnnotationsMutateWithAnnotations(t *testing.T) {
	want := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name: "test",
			Annotations: map[string]string{
				"test":    "true",
				"useless": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       "https",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			}},
		},
	}

	svc := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name: "test",
			Annotations: map[string]string{
				"useless": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       "https",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			}},
		},
	}

	got, err := InjectAnnotations{Logger: logger()}.Mutate(svc)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, want, got)
}
