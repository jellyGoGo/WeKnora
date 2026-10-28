package main

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestDesktopBackendWaitsForReadyURL(t *testing.T) {
	ready := make(chan struct{})
	failures := make(chan error, 1)
	finished := make(chan error, 1)
	go func() { finished <- waitForDesktopBackend(ready, failures, time.Second) }()
	select {
	case err := <-finished:
		t.Fatalf("returned before backend was ready: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	close(ready)
	if err := <-finished; err != nil {
		t.Fatal(err)
	}
}

func TestDesktopProxyServesAfterSlowBackendStartup(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "desktop frontend")
	}))
	defer backend.Close()
	ready := make(chan struct{})
	var backendURL string
	go func() {
		time.Sleep(600 * time.Millisecond)
		backendURL = backend.URL
		close(ready)
	}()
	if err := waitForDesktopBackend(ready, make(chan error), 2*time.Second); err != nil {
		t.Fatal(err)
	}
	target, err := url.Parse(backendURL)
	if err != nil {
		t.Fatal(err)
	}
	frontend := httptest.NewServer(desktopAPIProxy(target))
	defer frontend.Close()
	response, err := frontend.Client().Get(frontend.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || string(body) != "desktop frontend" {
		t.Fatalf("desktop proxy response = %d %q", response.StatusCode, body)
	}
}

func TestDesktopBackendStartupReportsFailureAndTimeout(t *testing.T) {
	failure := errors.New("backend initialization failed")
	failures := make(chan error, 1)
	failures <- failure
	if err := waitForDesktopBackend(make(chan struct{}), failures, time.Second); !errors.Is(err, failure) {
		t.Fatalf("startup error = %v, want original backend failure", err)
	}
	if err := waitForDesktopBackend(make(chan struct{}), make(chan error), time.Millisecond); err == nil {
		t.Fatal("backend that never becomes ready must time out")
	}
}
