package accesslog

import (
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/api"
)

func tamperService(svc *api.Service, config *accessLogConfig) error {

	if svc.Annotations != nil && svc.Annotations[AnnotationEnableKey] == "false" {
		glog.V(config.LogLevel).Infof("Skipping service %s/%s since explicitly annotation %s=false matched", svc.Namespace, svc.Name, AnnotationEnableKey)
		return nil
	}

	if svc.Spec.Selector == nil || svc.Spec.Selector[AppKey] == "" {
		glog.V(config.LogLevel).Infof("Skipping service %s/%s since it is missing mandatory label %s", svc.Namespace, svc.Name, AppKey)
		return nil
	}

	if svc.Spec.Type != api.ServiceTypeClusterIP {
		glog.V(config.LogLevel).Infof("Skipping service %s/%s since its not type of ClusterIP", svc.Namespace, svc.Name)
		return nil
	}

	if len(svc.Spec.Ports) == 0 {
		glog.V(config.LogLevel).Infof("Skipping service %s/%s since there is no existing ports", svc.Namespace, svc.Name)
		return nil
	}

	x := -1
	for i, p := range svc.Spec.Ports {
		if p.Port == 80 {
			x = i
			break
		}
	}
	if x == -1 {
		glog.V(config.LogLevel).Infof("Skipping service %s/%s since there is 80 port", svc.Namespace, svc.Name)
		return nil
	}

	nginxProxyPort := config.NginxSpec.Ports[0].ContainerPort
	svc.Spec.Ports[x].TargetPort = intstr.FromInt32(nginxProxyPort)
	glog.V(1).Infof("Service %s/%s targetPort has been changed to %d", svc.Namespace, svc.Name, nginxProxyPort)

	return nil
}
