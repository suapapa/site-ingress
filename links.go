package main

import (
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Link struct {
	Name         string `yaml:"name"`
	Link         string `yaml:"link"`
	Desc         string `yaml:"desc,omitempty"`
	ReverseProxy bool   `yaml:"reverse_proxy,omitempty"`
	Hide         bool   `yaml:"hide,omitempty"`
}

var (
	links                  []*Link
	lastLinksLoadTimeStamp time.Time
	redirects              = map[string]*Link{}
)

// update links every in 30 min interval
func updateLinks() error {
	if links == nil || time.Since(lastLinksLoadTimeStamp) > 30*time.Minute {
		ls, err := loadLinksConf(linksConf)
		if err != nil {
			return errors.Wrap(err, "fail to get links")
		}
		// update redirect map
		for k := range redirects {
			delete(redirects, k)
		}
		for i, l := range ls {
			redirects[l.Name] = ls[i]
		}

		lastLinksLoadTimeStamp = time.Now()
		links = ls
	}

	return nil
}

func loadLinksConf(path string) ([]*Link, error) {
	log.Println("load link conf")
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