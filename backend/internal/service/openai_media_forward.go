package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type OpenAIMediaProxyResult struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	RequestID  string
	Duration   time.Duration
}

func buildOpenAICompatibleURL(base, path string) string {
	normalized := strings.TrimRight(strings.TrimSpace(base), "/")
	path = strings.TrimSpace(path)
	if path == "" {
		return normalized
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if strings.HasPrefix(path, "/v1/") {
		return normalized + path
	}
	if strings.HasSuffix(normalized, "/v1") {
		return normalized + path
	}
	return normalized + "/v1" + path
}

func (s *OpenAIGatewayService) ForwardCompatibleMedia(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	method string,
	upstreamPath string,
	contentType string,
	body []byte,
) (*OpenAIMediaProxyResult, error) {
	startTime := time.Now()
	if s == nil || s.httpUpstream == nil {
		return nil, fmt.Errorf("http upstream client not configured")
	}
	token, _, err := s.GetAccessToken(ctx, account)
	if err != nil {
		return nil, err
	}

	baseURL := strings.TrimSpace(account.GetOpenAIBaseURL())
	if baseURL == "" {
		return nil, fmt.Errorf("base_url not found in credentials")
	}
	validatedURL, err := s.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return nil, err
	}
	upstreamURL := buildOpenAICompatibleURL(validatedURL, upstreamPath)

	upstreamReq, err := http.NewRequestWithContext(ctx, method, upstreamURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		upstreamReq.Header.Set("Content-Type", contentType)
	}
	upstreamReq.Header.Set("Authorization", "Bearer "+token)
	for _, header := range []string{"Accept", "Accept-Encoding"} {
		if c != nil {
			if value := c.GetHeader(header); value != "" {
				upstreamReq.Header.Set(header, value)
			}
		}
	}

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	resp, err := s.httpUpstream.Do(upstreamReq, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		return nil, &UpstreamFailoverError{StatusCode: http.StatusBadGateway}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if readErr != nil {
		return nil, readErr
	}

	if resp.StatusCode >= 400 && s.shouldFailoverUpstreamError(resp.StatusCode) {
		return nil, &UpstreamFailoverError{
			StatusCode:      resp.StatusCode,
			ResponseBody:    respBody,
			ResponseHeaders: resp.Header.Clone(),
		}
	}

	return &OpenAIMediaProxyResult{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header.Clone(),
		RequestID:  resp.Header.Get("x-request-id"),
		Duration:   time.Since(startTime),
	}, nil
}

func (s *OpenAIGatewayService) ProxyCompatibleMediaFile(
	ctx context.Context,
	account *Account,
	upstreamPath string,
) (*http.Response, error) {
	if s == nil || s.httpUpstream == nil {
		return nil, fmt.Errorf("http upstream client not configured")
	}
	baseURL := strings.TrimSpace(account.GetOpenAIBaseURL())
	if baseURL == "" {
		return nil, fmt.Errorf("base_url not found in credentials")
	}
	validatedURL, err := s.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return nil, err
	}
	upstreamURL := buildOpenAICompatibleURL(validatedURL, upstreamPath)

	upstreamReq, err := http.NewRequestWithContext(ctx, http.MethodGet, upstreamURL, nil)
	if err != nil {
		return nil, err
	}
	token, _, err := s.GetAccessToken(ctx, account)
	if err == nil && strings.TrimSpace(token) != "" {
		upstreamReq.Header.Set("Authorization", "Bearer "+token)
	}

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	return s.httpUpstream.Do(upstreamReq, proxyURL, account.ID, account.Concurrency)
}
