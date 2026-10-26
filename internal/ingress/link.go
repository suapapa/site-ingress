package ingress

import (
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
)

type Site struct {
	About About  `yaml:"about,omitempty"`
	Links Links  `yaml:"links"`
	Says  []*Say `yaml:"says"`
}

type About struct {
	Hero        string `yaml:"hero"`
	Name        string `yaml:"name"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

func (a About) IsEmpty() bool {
	return strings.TrimSpace(a.Hero) == "" &&
		strings.TrimSpace(a.Name) == "" &&
		strings.TrimSpace(a.Title) == "" &&
		strings.TrimSpace(a.Description) == ""
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
