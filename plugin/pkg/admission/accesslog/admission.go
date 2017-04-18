// Package accesslog inject nginx into deployment and modify service accordingly
package accesslog

import (
	"io"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/admission"
	"k8s.io/kubernetes/pkg/api"
	apierrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/apis/extensions"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/kubelet/util/sliceutils"
	"k8s.io/kubernetes/pkg/util/strings"
)

func init() {
	admission.RegisterPlugin("AccessLog", func(client clientset.Interface, config io.Reader) (admission.Interface, error) {
		return NewAccessLog(), nil
	})
	glog.V(6).Infof("Adimission controller AccessLog has been registered")
}

const (
	NginxProxyPort        = 80
	AppKey                = "app"
	AnnotationEnableKey   = "enable-access-log"
	AnnotationTargetKey   = "access-log-target"
	AnnotationTemplateKey = "access-log-template"
)

var (
	NamespaceWhitelist = []string{"cy", "xl", "ta-100"}
	NameBlacklist      = []string{"etcd", "redis", "eshop"}
)

// accessLog is an implementation of admission.Interface.
type accessLog struct {
	*admission.Handler
}

func (a *accessLog) Admit(attributes admission.Attributes) (err error) {

	glog.V(8).Infof("Begin of AccessLog:Admit: %s", attributes.GetKind().GroupKind())

	if !sliceutils.StringInSlice(attributes.GetNamespace(), NamespaceWhitelist) {
		glog.V(6).Infof("Drop namespace %s which is not in the whitelist", attributes.GetNamespace())
		return nil
	}

	if attributes.GetKind().GroupKind() == extensions.Kind("Deployment") {
		dep, ok := attributes.GetObject().(*extensions.Deployment)
		if !ok {
			return show(apierrors.NewBadRequest("Resource was marked with kind Deployment but was unable to be converted"))
		}
		if sliceutils.StringInSliceFunc(dep.Name, NameBlacklist, strings.Contains) {
			glog.V(6).Infof("Drop name %s which is in the blacklist", dep.Name)
			return nil
		}

		glog.V(6).Infof("Process deployment: %s/%s", dep.Namespace, dep.Name)
		if err := tamperDeployment(dep); err != nil {
			return show(err)
		}
	} else if attributes.GetKind().GroupKind() == api.Kind("Service") {
		svc, ok := attributes.GetObject().(*api.Service)
		if !ok {
			return show(apierrors.NewBadRequest("Resource was marked with kind Service but was unable to be converted"))
		}
		if sliceutils.StringInSliceFunc(svc.Name, NameBlacklist, strings.Contains) {
			glog.V(6).Infof("Drop name %s which is in the blacklist", svc.Name)
			return nil
		}

		glog.V(6).Infof("Process service: %s/%s", svc.Namespace, svc.Name)
		if err := tamperService(svc); err != nil {
			return show(err)
		}
	}

	glog.V(8).Infof("End of AccessLog:Admit")
	return nil
}

func NewAccessLog() admission.Interface {
	glog.V(6).Infof("Adimission controller AccessLog has been initialized")
	return &accessLog{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}
}

func show(e error) error {
	glog.V(1).Infof("AccessLog error: %s", e)
	return e
}
