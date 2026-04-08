package main

import (
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/pkg/errors"
)

type ReverseProxyCache struct {
	ReverseProxy *httputil.ReverseProxy
	URL          *url.URL
}

var (
	rpMu    sync.RWMutex
	rpCache = make(map[string]*ReverseProxyCache)
)

func getReverseProxy(target string) (*ReverseProxyCache, error) {
	rpMu.RLock()
	rpc, ok := rpCache[target]
	rpMu.RUnlock()
	if ok {
		return rpc, nil
	}

	rpMu.Lock()
	defer rpMu.Unlock()
	if rpc, ok := rpCache[target]; ok {
		return rpc, nil
	}

	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get reverse proxy")
	}

	rp := httputil.NewSingleHostReverseProxy(targetURL)
	rpc = &ReverseProxyCache{
		ReverseProxy: rp,
		URL:          targetURL,
	}
	rpCache[target] = rpc

	return rpc, nil
}
