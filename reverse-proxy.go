package main

import (
	"context"
	"net/http"
	"strings"
	// "go.opentelemetry.io/otel/attribute"
)

// Serve a reverse proxy for a given url
func serveReverseProxy(ctx context.Context, res http.ResponseWriter, req *http.Request, from, to string) {
	// _, span := tracer.Start(ctx, "serve-reverse-proxy")
	// defer span.End()

	rpc, err := getReverseProxy(from)
	if err != nil {
		log.Errorf("fail serve reverse proxy: %v", err)
	}
	// span.SetAttributes(
	// 	attribute.String("from", from),
	// 	attribute.String("to", to),
	// )

	url, proxy := rpc.URL, rpc.ReverseProxy

	// Create a custom director function to handle the request
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = url.Scheme
		req.URL.Host = url.Host
		req.URL.Path = to
		req.Host = url.Host

		// Set headers to prevent redirect loops
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Header.Set("X-Forwarded-Proto", req.URL.Scheme)
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)

		// Log request details
		log.Infof("proxy request: method=%s, url=%s, host=%s", req.Method, req.URL.String(), req.Host)
	}

	// Add error handler to prevent redirect loops
	proxy.ErrorHandler = func(res http.ResponseWriter, req *http.Request, err error) {
		log.Errorf("proxy error: %v", err)
		http.Error(res, "Proxy Error", http.StatusBadGateway)
	}

	// Add modify response to handle redirects
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Log response details
		log.Infof("proxy response: status=%d, headers=%v", resp.StatusCode, resp.Header)

		// If we get a redirect response, modify it to prevent loops
		if resp.StatusCode >= 300 && resp.StatusCode <= 399 {
			location := resp.Header.Get("Location")
			if location != "" {
				log.Infof("redirect detected: from=%s to=%s", req.URL.Path, location)

				// Handle both absolute and relative path redirects
				if location == req.URL.Path || location == req.URL.Path+"/" ||
					location == strings.TrimSuffix(req.URL.Path, "/") ||
					location == strings.TrimPrefix(req.URL.Path, "/") {
					log.Infof("preventing redirect loop: path=%s", location)
					resp.StatusCode = http.StatusOK
					resp.Header.Del("Location")
				}
			}
		}
		return nil
	}

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}
