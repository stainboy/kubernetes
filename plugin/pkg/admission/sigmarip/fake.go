package sigmarip

import (
	"k8s.io/apiserver/pkg/admission"
)

type fake struct {
	*admission.Handler
}

func (a *fake) Admit(attributes admission.Attributes) (err error) {
	return nil
}
