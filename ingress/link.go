package ingress

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Link struct {
	Name string `yaml:"name"`
	Link string `yaml:"link"`
	Desc string `yaml:"desc,omitempty"`
	Hide bool   `yaml:"hide,omitempty"`
	// for rever-proxy
	RPLink string `yaml:"reverse_proxy_link,omitempty"`
	// for port-foward
	PortFoward bool `yaml:"port_foward,omitempty"`
	InPort     int  `yaml:"in_port,omitempty"`
	OutPort    int  `yaml:"out_port,omitempty"`
}

func LoadLinksConf(path string) ([]*Link, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "fail to load links conf")
	}

	var ret []*Link
	err = yaml.NewDecoder(f).Decode(&ret)
	if err != nil {
		return nil, errors.Wrap(err, "fail to load links conf")
	}

	return ret, nil
}
