package sigmarip

import (
	"io"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/util/yaml"
)

// sigmaRipConfig holds config data for imagePolicyWebhook
type sigmaRipConfig struct {
	Enabled  bool       `json:"enabled"`
	LogLevel glog.Level `json:"log"`
	Target   string     `json:"target"`
}

// AdmissionConfig holds config data for admission controllers
type AdmissionConfig struct {
	SigmaRip sigmaRipConfig `json:"sigmaRip"`
}

func newConfig(s io.Reader) (*sigmaRipConfig, error) {
	var config AdmissionConfig
	d := yaml.NewYAMLOrJSONDecoder(s, 4096)
	err := d.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config.SigmaRip, nil
}
