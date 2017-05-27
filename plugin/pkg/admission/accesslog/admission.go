// Package accesslog inject nginx into deployment and modify service accordingly
package accesslog

import (
	"io"
	"strings"

	"github.com/golang/glog"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

func init() {
	admission.RegisterPlugin("AccessLog", func(config io.Reader) (admission.Interface, error) {
		return NewAccessLog(config)
	})
	glog.V(1).Infof("Adimission controller AccessLog has been registered")
}

const (
	AppKey                = "app"
	AnnotationEnableKey   = "enable-access-log"
	AnnotationTargetKey   = "access-log-target"
	AnnotationTemplateKey = "access-log-template"
)

// accessLog is an implementation of admission.Interface.
type accessLog struct {
	Config *accessLogConfig
	*admission.Handler
}

func (a *accessLog) Admit(attributes admission.Attributes) (err error) {
	glog.V(a.Config.LogLevel).Infof("Begin of AccessLog:Admit: %s", attributes.GetKind().GroupKind())

	if !test(attributes.GetNamespace(), &a.Config.Namespace) {
		glog.V(a.Config.LogLevel).Infof("Drop namespace %s which doesn't match black-white-list", attributes.GetNamespace())
		return nil
	}

	if attributes.GetKind().GroupKind() == extensions.Kind("Deployment") {
		dep, ok := attributes.GetObject().(*extensions.Deployment)
		if !ok {
			return show(apierrors.NewBadRequest("Resource was marked with kind Deployment but was unable to be converted"))
		}
		if !test(dep.Name, &a.Config.Deployment) {
			glog.V(a.Config.LogLevel).Infof("Drop name %s which doesn't match black-white-list", dep.Name)
			return nil
		}

		glog.V(a.Config.LogLevel).Infof("Process deployment: %s/%s", dep.Namespace, dep.Name)
		if err := tamperDeployment(dep, a.Config); err != nil {
			return show(err)
		}
	} else if attributes.GetKind().GroupKind() == api.Kind("Service") {
		svc, ok := attributes.GetObject().(*api.Service)
		if !ok {
			return show(apierrors.NewBadRequest("Resource was marked with kind Service but was unable to be converted"))
		}
		if !test(svc.Name, &a.Config.Service) {
			glog.V(a.Config.LogLevel).Infof("Drop name %s which doesn't match black-white-list", svc.Name)
			return nil
		}

		glog.V(a.Config.LogLevel).Infof("Process service: %s/%s", svc.Namespace, svc.Name)
		if err := tamperService(svc, a.Config); err != nil {
			return show(err)
		}
	}

	glog.V(a.Config.LogLevel).Infof("End of AccessLog:Admit")
	return nil
}

func NewAccessLog(s io.Reader) (admission.Interface, error) {

	config, err := newConfig(s)
	if err != nil {
		return nil, err
	}

	glog.V(1).Infof("Adimission controller AccessLog(%v) has been initialized", config.Enabled)
	if config.Enabled {
		return &accessLog{
			Config:  config,
			Handler: admission.NewHandler(admission.Create, admission.Update),
		}, nil
	}

	return &fake{admission.NewHandler()}, nil
}

func show(e error) error {
	glog.V(1).Infof("AccessLog error: %s", e)
	return e
}

func test(s string, list *blackWhiteList) bool {
	for _, r := range list.Exclude {
		if r == "*" || strings.Contains(s, r) {
			return false
		}
	}
	for _, r := range list.Include {
		if r == "*" || strings.Contains(s, r) {
			return true
		}
	}
	return false
}
