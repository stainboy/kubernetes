package accesslog

import (
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/api"
)

func tamperService(svc *api.Service) error {

	if svc.Annotations[AnnotationEnableKey] == "false" {
		return nil
	}

	if svc.Spec.Type != api.ServiceTypeLoadBalancer {
		return nil
	}

	if svc.Labels[AppKey] == "" {
		return nil
	}

	if len(svc.Spec.Ports) == 0 {
		return nil
	}

	for _, p := range svc.Spec.Ports {
		if p.Port == 80 {
			p.TargetPort = intstr.FromInt(NginxProxyPort)
		}
	}

	return nil
}
