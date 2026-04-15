package common

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

// NewHTTPClient creates an HTTP client that supports proxy configuration
// via HTTP_PROXY and HTTPS_PROXY environment variables.
func NewHTTPClient() *http.Client {
	proxyURL := getProxyURL()

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   5 * time.Minute,
	}
}

// getProxyURL returns the proxy URL based on environment variables.
// It checks HTTPS_PROXY for HTTPS requests and HTTP_PROXY for HTTP requests.
func getProxyURL() *url.URL {
	// Check for HTTPS proxy first (used for HTTPS requests)
	httpsProxy := os.Getenv("HTTPS_PROXY")
	if httpsProxy != "" {
		if proxyURL, err := url.Parse(httpsProxy); err == nil {
			return proxyURL
		}
	}

	// Check for HTTP proxy (used for HTTP requests or fallback)
	httpProxy := os.Getenv("HTTP_PROXY")
	if httpProxy != "" {
		if proxyURL, err := url.Parse(httpProxy); err == nil {
			return proxyURL
		}
	}

	return nil
}
