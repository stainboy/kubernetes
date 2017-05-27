package sigmarip

import (
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

// TestAdmission verifies ...
func TestAdmission(t *testing.T) {
	dep := extensions.Deployment{
		Spec: extensions.DeploymentSpec{
			Template: api.PodTemplateSpec{
				Spec: api.PodSpec{
					Containers: []api.Container{
						api.Container{
							Name: "bss",
						},
						api.Container{
							Name: "monitor",
						},
						api.Container{
							Name: "nginx",
						},
					},
				},
			},
		},
	}

	config := sigmaRipConfig{
		Target: "monitor",
	}

	err := tamperDeployment(&dep, &config)
	t.Log(err)

	t.Log(len(dep.Spec.Template.Spec.Containers))
}
