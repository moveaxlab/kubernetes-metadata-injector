package mutation

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestMutateServicePatch(t *testing.T) {
	m := NewMutator(logger())
	got, err := m.MutateServicePatch(service())
	if err != nil {
		t.Fatal(err)
	}

	p := servicePatch()
	g := string(got)
	assert.Equal(t, p, g)
}

func service() *corev1.Service {
	return &corev1.Service{
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
}

func servicePatch() string {
	patch := `[
		{"op":"add","path":"/metadata/annotations/test","value": "true"}
]`

	patch = strings.ReplaceAll(patch, "\n", "")
	patch = strings.ReplaceAll(patch, "\t", "")
	patch = strings.ReplaceAll(patch, " ", "")

	return patch
}

func logger() *logrus.Entry {
	mute := logrus.StandardLogger()
	mute.Out = ioutil.Discard
	return mute.WithField("logger", "test")
}
