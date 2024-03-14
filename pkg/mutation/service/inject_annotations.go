package service

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// injectAnnotations is a container for the mutation injecting environment vars
type InjectAnnotations struct {
	Logger logrus.FieldLogger
}

// injectAnnotations implements the ServiceMutator interface
var _ ServiceMutator = (*InjectAnnotations)(nil)

// Name returns the struct name
func (se InjectAnnotations) Name() string {
	return "inject_annotations"
}

// Mutate returns a new mutated service according to set env rules
func (se InjectAnnotations) Mutate(service *corev1.Service) (*corev1.Service, error) {
	se.Logger = se.Logger.WithField("mutation", se.Name())
	mservice := service.DeepCopy()
	triggerAnnotationPrefix := os.Getenv("TRIGGER_ANNOTATION_PREFIX") + ".annotation."
	if triggerAnnotationPrefix == "" {
		se.Logger.Fatal("env var TRIGGER_ANNOTATION_PREFIX not set")
		os.Exit(1)
	}

	if service.Annotations != nil {

		annotations := se.getRelatedAnnotations(service.Annotations, triggerAnnotationPrefix)

		if len(annotations) != 0 {
			// inject annotaions into service
			for key, value := range annotations {
				secretNamespace, secertName, secretKey := se.getSecretRef(value)
				if secretNamespace != service.Namespace {
					// avoid accessing information in different namespace
					return nil, fmt.Errorf("Security error: service '%s/%s' is trying to reference secret in differnt namespace: '%s/%s'", service.Namespace, service.Name, secretNamespace, secertName)

				}
				annotationValue, err := se.getSecretValue(secretNamespace, secertName, secretKey)
				if err != nil {
					return nil, err
				}

				se.injectAnnotation(mservice, key, annotationValue)
				se.Logger.Debugf("service annotation injected '%s: %s'", key, annotationValue)

			}
			se.Logger.Infof("Injected annotations for service '%s/%s'", service.Namespace, service.Name)
		} else {
			se.Logger.Debugf("Ignored service '%s/%s' no trigger annotation prefix found", service.Namespace, service.Name)
		}
	}
	return mservice, nil
}

// injectAnnotation injects a single annotation overrriding if already present (to allow UPDATE)
func (se InjectAnnotations) injectAnnotation(service *corev1.Service, annotationKey string, annotationValue string) {
	ann := service.Annotations
	if ann == nil {
		service.Annotations = make(map[string]string)
	}
	service.Annotations[annotationKey] = annotationValue
}

var k8s *kubernetes.Clientset

func k8sClient() *kubernetes.Clientset {
	if k8s == nil {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}
		k8s = clientset
	}
	return k8s
}

func (se InjectAnnotations) getRelatedAnnotations(annotations map[string]string, triggerAnnotationPrefix string) map[string]string {
	// related annotation key are in the forma "TRIGGER_ANNOTATION_PREFIX.annotation.actual-annotation"
	relatedAnnotations := make(map[string]string)
	for a_k, a_v := range annotations {
		if strings.HasPrefix(a_k, triggerAnnotationPrefix) {
			actualAnnotation := strings.TrimPrefix(a_k, triggerAnnotationPrefix)
			relatedAnnotations[actualAnnotation] = a_v
		}
	}

	return relatedAnnotations
}

func (se InjectAnnotations) getSecretRef(annotationValue string) (string, string, string) {
	// value must be in the format "secret-namespace/secret-name:secret-key"
	split := strings.Split(annotationValue, "/")
	secretNamespace := split[0]
	split1 := strings.Split(split[1], ":")
	secertName := split1[0]
	secretKey := split1[1]

	return secretNamespace, secertName, secretKey

}

func (se InjectAnnotations) getSecretValue(secretNamespace string, secertName string, secretKey string) (string, error) {
	secretObject, err := k8sClient().CoreV1().Secrets(secretNamespace).Get(context.Background(), secertName, metav1.GetOptions{})
	if err != nil {
		se.Logger.Debugf("Cannot access secret '%s/%s'", secretNamespace, secertName)
		return "", err
	}
	value, ok := secretObject.Data[secretKey]
	if !ok {
		se.Logger.Debugf("secret '%s/%s' has not key '%s'", secretNamespace, secertName, secretKey)
		return "", fmt.Errorf("secret '%s/%s' has not key '%s'", secretNamespace, secertName, secretKey)
	}
	return string(value), nil
}
