package accesslog

import (
	"io"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/util/yaml"
)

// accessLogConfig holds config data for imagePolicyWebhook
type accessLogConfig struct {
	Enabled  bool       `json:"enabled"`
	LogLevel glog.Level `json:"log"`

	Namespace  blackWhiteList `json:"namespace"`
	Deployment blackWhiteList `json:"deployment"`
	Service    blackWhiteList `json:"service"`
	NginxSpec  api.Container  `json:"spec"`
}

type blackWhiteList struct {
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`

	// IncludeRegex []*regexp.Regexp
	// ExcludeRegex []*regexp.Regexp
}

// admissionConfig holds config data for admission controllers
type admissionConfig struct {
	AccessLog accessLogConfig `json:"accessLog"`
}

func newConfig(s io.Reader) (*accessLogConfig, error) {
	var config admissionConfig
	d := yaml.NewYAMLOrJSONDecoder(s, 4096)
	err := d.Decode(&config)
	if err != nil {
		return nil, err
	}

	// err = compileRegex(&config.AccessLog)
	// if err != nil {
	// 	return nil, err
	// }

	return &config.AccessLog, nil
}

// func compileRegex(config *accessLogConfig) error {

// 	for _, l := range []*blackWhiteList{&config.Namespace, &config.Deployment, &config.Service} {
// 		l.IncludeRegex = make([]*regexp.Regexp, len(l.Include))
// 		for _, s := range l.Include {
// 			if r, err := regexp.Compile(s); err != nil {
// 				return err
// 			} else {
// 				r.MatchString("") // create regexp cache to avoid crashing during multi-thread call
// 				l.IncludeRegex = append(l.IncludeRegex, r)
// 			}
// 		}

// 		l.ExcludeRegex = make([]*regexp.Regexp, len(l.Exclude))
// 		for _, s := range l.Exclude {
// 			if r, err := regexp.Compile(s); err != nil {
// 				return err
// 			} else {
// 				r.MatchString("") // create regexp cache to avoid crashing during multi-thread call
// 				l.ExcludeRegex = append(l.ExcludeRegex, r)
// 			}
// 		}
// 	}

// 	return nil
// }
