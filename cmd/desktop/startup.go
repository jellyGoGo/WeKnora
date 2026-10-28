package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func waitForDesktopBackend(ready <-chan struct{}, failures <-chan error, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-ready:
		return nil
	case err := <-failures:
		return fmt.Errorf("desktop backend startup: %w", err)
	case <-timer.C:
		return fmt.Errorf("desktop backend did not become ready within %s", timeout)
	}
}

// desktopAPIProxy forwards the webview to the loopback Gin server. Pairing
// needs the API listener as Host so the link cannot be aimed at another
// machine. Other routes keep the page Host (wails.localhost), which OIDC
// callbacks and embed origin checks still compare against.
func desktopAPIProxy(target *url.URL) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			r.SetXForwarded()
			if !isBrowserPairingRequest(r.In) {
				r.Out.Host = r.In.Host
			}
		},
	}
}

func isBrowserPairingRequest(req *http.Request) bool {
	return req != nil && req.URL != nil && req.Method == http.MethodPost && req.URL.Path == "/api/v1/me/browser"
}
