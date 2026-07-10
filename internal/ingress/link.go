package ingress

import (
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

type Links map[string][]*Link

type Link struct {
	Name    string `yaml:"name"`
	Link    string `yaml:"link"`
	Desc    string `yaml:"desc,omitempty"`
	Hide    bool   `yaml:"hide,omitempty"`
	SiteMap bool   `yaml:"site_map,omitempty"`
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
