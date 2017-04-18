// Package sigmarip inject nginx into deployment and modify service accordingly
package sigmarip

import (
	"io"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/admission"
	apierrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/apis/extensions"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
)

func init() {
	admission.RegisterPlugin("SigmaRip", func(client clientset.Interface, config io.Reader) (admission.Interface, error) {
		return NewSigmaRip(config)
	})
	glog.V(1).Infof("Adimission controller SigmaRip has been registered")
}

// sigmaRip is an implementation of admission.Interface.
type sigmaRip struct {
	Config *sigmaRipConfig
	*admission.Handler
}

func (a *sigmaRip) Admit(attributes admission.Attributes) (err error) {
	if !a.Config.Enabled {
		return nil
	}
	glog.V(a.Config.LogLevel).Infof("Begin of SigmaRip:Admit: %s", attributes.GetKind().GroupKind())

	if attributes.GetKind().GroupKind() == extensions.Kind("Deployment") {
		dep, ok := attributes.GetObject().(*extensions.Deployment)
		if !ok {
			return show(apierrors.NewBadRequest("Resource was marked with kind Deployment but was unable to be converted"))
		}

		glog.V(a.Config.LogLevel).Infof("Process deployment: %s/%s", dep.Namespace, dep.Name)
		if err := tamperDeployment(dep, a.Config); err != nil {
			return show(err)
		}
	}

	glog.V(a.Config.LogLevel).Infof("End of SigmaRip:Admit")
	return nil
}

func NewSigmaRip(s io.Reader) (admission.Interface, error) {

	config, err := newConfig(s)
	if err != nil {
		return nil, err
	}

	glog.V(1).Infof("Adimission controller SigmaRip has been initialized")
	return &sigmaRip{
		Config:  config,
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}, nil
}

func show(e error) error {
	glog.V(1).Infof("SigmaRip error: %s", e)
	return e
}
