package ingress

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
)

type Site struct {
	Links Links  `yaml:"links"`
	Says  []*Say `yaml:"says"`
}

type Say struct {
	Movie     string `yaml:"movie"`
	Line      string `yaml:"line"`
	Character string `yaml:"character"`
}

func (ml *Say) String() string {
	return fmt.Sprintf("“%s” - %s; %s", ml.Line, ml.Character, ml.Movie)
}

type Links map[string][]*Link

type Link struct {
	Name string `yaml:"name"`
	Link string `yaml:"link"`
	Desc string `yaml:"desc,omitempty"`
	Hide bool   `yaml:"hide,omitempty"`
	// for rever-proxy
	RPLink       string `yaml:"reverse_proxy_link,omitempty"`
	RPOmitPrefix bool   `yaml:"reverse_proxy_omit_prefix,omitempty"`
	// for port-foward
	PortFoward bool `yaml:"port_foward,omitempty"`
	InPort     int  `yaml:"in_port,omitempty"`
	OutPort    int  `yaml:"out_port,omitempty"`

	SiteMap bool `yaml:"site_map,omitempty"`
}

func LoadSiteFromFile(path string) (*Site, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "fail to load site conf")
	}
	defer f.Close()

	ret := &Site{}
	err = yaml.NewDecoder(f).Decode(ret)
	if err != nil {
		return nil, errors.Wrap(err, "fail to load site conf")
	}

	return ret, nil
}
