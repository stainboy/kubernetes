package sigmarip

import (
	"k8s.io/kubernetes/pkg/apis/extensions"
)

func tamperDeployment(dep *extensions.Deployment, config *sigmaRipConfig) error {

	x := -1
	for i, c := range dep.Spec.Template.Spec.Containers {
		if c.Name == config.Target {
			x = i
			break
		}
	}

	if x != -1 {
		// delete the specific container
		c := dep.Spec.Template.Spec.Containers
		c = append(c[:x], c[x+1:]...)
		dep.Spec.Template.Spec.Containers = c
	}

	return nil
}
