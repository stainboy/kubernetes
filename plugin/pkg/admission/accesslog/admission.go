// Package accesslog inject nginx into deployment and modify service accordingly
package accesslog

import (
	"io"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

func init() {
	admission.RegisterPlugin("AccessLog", func(config io.Reader) (admission.Interface, error) {
		return NewAccessLog(), nil
	})
}

const (
	NginxProxyPort        = 8038
	AppKey                = "app"
	AnnotationEnableKey   = "enable-access-log"
	AnnotationTargetKey   = "access-log-target"
	AnnotationTemplateKey = "access-log-template"
)

// accessLog is an implementation of admission.Interface.
type accessLog struct {
	*admission.Handler
}

func (a *accessLog) Admit(attributes admission.Attributes) (err error) {

	if len(attributes.GetSubresource()) != 0 {
		return nil
	} else if attributes.GetResource().GroupResource() == api.Resource("deployments") {
		dep, ok := attributes.GetObject().(*extensions.Deployment)
		if !ok {
			return apierrors.NewBadRequest("Resource was marked with kind Deployment but was unable to be converted")
		}
		if err := tamperDeployment(dep); err != nil {
			return err
		}
	} else if attributes.GetResource().GroupResource() == api.Resource("services") {
		svc, ok := attributes.GetObject().(*api.Service)
		if !ok {
			return apierrors.NewBadRequest("Resource was marked with kind Service but was unable to be converted")
		}
		if err := tamperService(svc); err != nil {
			return err
		}
	}

	return nil
}

func NewAccessLog() admission.Interface {
	return &accessLog{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}
}
