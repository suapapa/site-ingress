package main

import (
	"github.com/pkg/errors"
	"github.com/suapapa/site-ingress/ingress"
)

var (
	gLinks []*ingress.Link
)

// update links every in 30 min interval
func getLinks() ([]*ingress.Link, error) {
	if gLinks != nil {
		return gLinks, nil
	}

	ls, err := ingress.LoadLinksConf(linksConf)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get links")
	}
	gLinks = ls

	return gLinks, nil
}
