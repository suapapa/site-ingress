package main

import (
	"net/http/httputil"
	"net/url"

	"github.com/pkg/errors"
)

type ReverseProxyCache struct {
	ReverseProxy *httputil.ReverseProxy
	URL          *url.URL
}

var (
	rpCache map[string]*ReverseProxyCache
)

func init() {
	rpCache = map[string]*ReverseProxyCache{}
}

func getReverseProxy(target string) (*ReverseProxyCache, error) {
	if rpc, ok := rpCache[target]; ok {
		return rpc, nil
	}

	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get reverse proxy")
	}

	rp := httputil.NewSingleHostReverseProxy(targetURL)
	rpc := &ReverseProxyCache{
		ReverseProxy: rp,
		URL:          targetURL,
	}
	rpCache[target] = rpc

	return rpc, nil
}
