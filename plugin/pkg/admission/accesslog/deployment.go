package accesslog

import (
	"fmt"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

func tamperDeployment(dep *extensions.Deployment, config *accessLogConfig) error {

	// 1. validate whether the given deployment is eligible to attach nginx
	if !eligible(dep, config) {
		return nil
	}

	// 2. analyize default container
	targetContainer := analyzeDefaultContainer(dep, config)
	if len(targetContainer.Ports) == 0 {
		return nil
	}

	// 3. append container nginx
	nginx := encapsulateNginx(targetContainer, config)
	dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers, nginx)
	glog.V(config.LogLevel).Infof("Deployment %s/%s has been tempered", dep.Namespace, dep.Name)

	return nil
}

func eligible(dep *extensions.Deployment, config *accessLogConfig) bool {

	// 1. check annotations.enable-access-log
	if dep.Annotations != nil && dep.Annotations[AnnotationEnableKey] == "false" {
		glog.V(config.LogLevel).Infof("Skipping deployment %s/%s since explicitly annotation %s=false matched", dep.Namespace, dep.Name, AnnotationEnableKey)
		return false
	}

	// 2. test labels app={target}
	if dep.Spec.Template.Labels == nil || dep.Spec.Template.Labels[AppKey] == "" {
		glog.V(config.LogLevel).Infof("Skipping deployment %s/%s since it is missing mandatory label %s", dep.Namespace, dep.Name, AppKey)
		return false
	}

	if len(dep.Spec.Template.Spec.Containers) == 0 {
		glog.V(config.LogLevel).Infof("Skipping deployment %s/%s since there is no container", dep.Namespace, dep.Name)
		return false
	}

	// 3. test whether container `nginx` already exists or 80 port is used
	nginxProxyPort := config.NginxSpec.Ports[0].ContainerPort
	for _, c := range dep.Spec.Template.Spec.Containers {
		if c.Name == "nginx" {
			glog.V(config.LogLevel).Infof("Skipping deployment %s/%s since explicitly container=nginx found", dep.Namespace, dep.Name)
			return false
		}
		for _, p := range c.Ports {
			if p.ContainerPort == nginxProxyPort {
				glog.V(config.LogLevel).Infof("Skipping deployment %s/%s since expected container port %d was occupied by %s", dep.Namespace, dep.Name, nginxProxyPort, c.Name)
				return false
			}
		}
	}

	return true
}

func analyzeDefaultContainer(dep *extensions.Deployment, config *accessLogConfig) *api.Container {
	if len(dep.Spec.Template.Spec.Containers) == 1 {
		return &dep.Spec.Template.Spec.Containers[0]
	}

	if dep.Spec.Template.Labels != nil {
		if app := dep.Spec.Template.Labels[AppKey]; app != "" {
			for _, c := range dep.Spec.Template.Spec.Containers {
				if c.Name == app {
					return &c
				}
			}
		}
	}

	return &dep.Spec.Template.Spec.Containers[0]
}

func encapsulateNginx(target *api.Container, config *accessLogConfig) api.Container {
	// copy spec to local variable
	nginx := config.NginxSpec
	nginx.Env[0].Value = encapsulateNginxArgs(target)
	return nginx
}

func encapsulateNginxArgs(target *api.Container) string {
	return fmt.Sprintf(`{"target":{"name":"%s","port":%d}}`, target.Name, target.Ports[0].ContainerPort)
}
