package httpclient

import (
	"net/http"
	"sync"
	"time"
)

var (
	defaultClient     *http.Client
	defaultClientOnce sync.Once
)

// GetDefaultClient returns a shared HTTP client with connection pooling
// This client is thread-safe and should be reused across requests
func GetDefaultClient() *http.Client {
	defaultClientOnce.Do(func() {
		defaultClient = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,              // 最大空闲连接数
				MaxIdleConnsPerHost: 10,               // 每个host的最大空闲连接
				MaxConnsPerHost:     20,               // 每个host的最大连接数
				IdleConnTimeout:     90 * time.Second, // 空闲连接超时
				DisableCompression:  false,
				DisableKeepAlives:   false,
			},
		}
	})
	return defaultClient
}

// GetClientWithTimeout returns a new HTTP client with custom timeout
// For frequently used clients, consider creating a package-level variable instead
func GetClientWithTimeout(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     20,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
			DisableKeepAlives:   false,
		},
	}
}
