package services

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

type HTTPClientService struct {
	cache map[string]*http.Client
	mu    sync.RWMutex
}

func NewHTTPClientService() *HTTPClientService {
	return &HTTPClientService{
		cache: make(map[string]*http.Client),
	}
}

func (s *HTTPClientService) GetClient(proxyURL string) *http.Client {
	cacheKey := proxyURL

	s.mu.RLock()
	if client, ok := s.cache[cacheKey]; ok {
		s.mu.RUnlock()
		return client
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if client, ok := s.cache[cacheKey]; ok {
		return client
	}

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	if proxyURL != "" {
		transport := s.buildProxyTransport(proxyURL)
		if transport != nil {
			client.Transport = transport
		}
	}

	s.cache[cacheKey] = client
	return client
}

func (s *HTTPClientService) buildProxyTransport(proxyURL string) *http.Transport {
	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil
	}

	var transport *http.Transport

	switch parsed.Scheme {
	case "http", "https":
		transport = &http.Transport{
			Proxy: http.ProxyURL(parsed),
		}
	case "socks5":
		var auth *proxy.Auth
		if parsed.User != nil {
			username := parsed.User.Username()
			password, _ := parsed.User.Password()
			auth = &proxy.Auth{User: username, Password: password}
		}

		dialer, err := proxy.SOCKS5("tcp", parsed.Host, auth, proxy.Direct)
		if err != nil {
			return nil
		}

		transport = &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			},
		}
	}

	return transport
}
