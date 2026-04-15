package common

import (
    "net"
    "net/http"
    "time"
)

// NewHTTPClient creates an HTTP client that supports proxy configuration
// via HTTP_PROXY and HTTPS_PROXY environment variables.
func NewHTTPClient() *http.Client {

    transport := &http.Transport{
        Proxy: http.ProxyFromEnvironment,
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
