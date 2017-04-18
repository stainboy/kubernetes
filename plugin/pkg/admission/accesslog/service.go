package accesslog

import (
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/util/intstr"
)

func tamperService(svc *api.Service) error {

	if svc.Annotations != nil && svc.Annotations[AnnotationEnableKey] == "false" {
		glog.V(6).Infof("Skipping service %s/%s since explicitly annotation %s=false matched", svc.Namespace, svc.Name, AnnotationEnableKey)
		return nil
	}

	if svc.Spec.Selector == nil || svc.Spec.Selector[AppKey] == "" {
		glog.V(6).Infof("Skipping service %s/%s since it is missing mandatory label %s", svc.Namespace, svc.Name, AppKey)
		return nil
	}

	if svc.Spec.Type != api.ServiceTypeClusterIP {
		glog.V(6).Infof("Skipping service %s/%s since its not type of ClusterIP", svc.Namespace, svc.Name)
		return nil
	}

	if len(svc.Spec.Ports) == 0 {
		glog.V(6).Infof("Skipping service %s/%s since there is no existing ports", svc.Namespace, svc.Name)
		return nil
	}

	x := 0
	for i, p := range svc.Spec.Ports {
		if p.Port == 80 {
			x = i
			break
		}
	}
	svc.Spec.Ports[x].TargetPort = intstr.FromInt(NginxProxyPort)
	glog.V(1).Infof("Service %s/%s targetPort has been changed to %d", svc.Namespace, svc.Name, NginxProxyPort)

	return nil
}
