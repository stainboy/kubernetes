package accesslog

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

func tamperDeployment(dep *extensions.Deployment) error {

	// 1. validate whether the given deployment is eligible to attach nginx
	if !eligible(dep) {
		return nil
	}

	// 2. analyize default container
	targetContainer := analyzeDefaultContainer(dep)
	if len(targetContainer.Ports) == 0 {
		return nil
	}

	// 3. append container nginx
	nginx := encapsulateNginx(targetContainer)
	dep.Spec.Template.Annotations[AnnotationTargetKey] = encapsulateNginxArgs(targetContainer)
	dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers, nginx)

	return nil
}

func eligible(dep *extensions.Deployment) bool {

	// 1. check annotations.enable-access-log
	if dep.Annotations[AnnotationEnableKey] == "false" {
		return false
	}

	// 2. test whether container `nginx` already exists
	for _, c := range dep.Spec.Template.Spec.Containers {
		if c.Name == "nginx" {
			return false
		}
	}

	// 3. test labels app={target}
	if dep.Spec.Template.Labels[AppKey] == "" {
		return false
	}

	if len(dep.Spec.Template.Spec.Containers) == 0 {
		return false
	}

	return true
}

func analyzeDefaultContainer(dep *extensions.Deployment) api.Container {
	if len(dep.Spec.Template.Spec.Containers) == 1 {
		return dep.Spec.Template.Spec.Containers[0]
	}

	app := dep.Spec.Template.Labels[AppKey]
	for _, c := range dep.Spec.Template.Spec.Containers {
		if c.Name == app {
			return c
		}
	}

	return dep.Spec.Template.Spec.Containers[0]
}

func encapsulateNginx(target api.Container) api.Container {
	return api.Container{
		Name:  "nginx",
		Image: "hyper.cd/occ/nginx-access-log:latest",
		Ports: []api.ContainerPort{
			api.ContainerPort{
				ContainerPort: NginxProxyPort,
			},
		},
		Resources: api.ResourceRequirements{
			Limits: api.ResourceList{
				api.ResourceCPU:    *resource.NewMilliQuantity(300, resource.DecimalSI),
				api.ResourceMemory: *resource.NewQuantity(48*1024*1024, resource.BinarySI),
			},
			Requests: api.ResourceList{
				api.ResourceCPU: *resource.NewMilliQuantity(50, resource.DecimalSI),
			},
		},
		Env: []api.EnvVar{
			api.EnvVar{
				Name: "TRACE_TARGET",
				ValueFrom: &api.EnvVarSource{
					FieldRef: &api.ObjectFieldSelector{
						FieldPath: fmt.Sprintf("metadata.annotations.%s", AnnotationTargetKey),
					},
				},
			},
			api.EnvVar{
				Name: "TRACE_TEMPLATE",
				ValueFrom: &api.EnvVarSource{
					FieldRef: &api.ObjectFieldSelector{
						FieldPath: fmt.Sprintf("metadata.annotations.%s", AnnotationTemplateKey),
					},
				},
			},
		},
	}
}

func encapsulateNginxArgs(target api.Container) string {
	return fmt.Sprintf("%s:%d", target.Name, target.Ports[0].ContainerPort)
}
