package main

import (
	"github.com/pkg/errors"
	"github.com/suapapa/site-ingress/ingress"
)

// update links every in 30 min interval
func getLinks(linksConfFile string) ([]*ingress.Link, error) {
	ls, err := ingress.LoadLinksConf(linksConfFile)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get links")
	}

	return ls, nil
}
