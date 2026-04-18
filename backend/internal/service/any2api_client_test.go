package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newAny2APIProxyTestContext(body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("User-Agent", "every2api-test")
	c.Request = req
	return c, rec
}

func TestAny2APIClientProxyRequest_ParsesResponsesUsageFromStandardSSE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, strings.Join([]string{
			"event: response.created",
			`data: {"type":"response.created","response":{"id":"resp_1"}}`,
			"",
			"event: response.completed",
			`data: {"type":"response.completed","response":{"id":"resp_1","usage":{"input_tokens":11,"output_tokens":7,"input_tokens_details":{"cached_tokens":3}}}}`,
			"",
			"data: [DONE]",
			"",
		}, "\n"))
	}))
	defer server.Close()

	client := &Any2APIClient{enabled: true, baseURL: server.URL, apiKey: "test-key"}
	c, rec := newAny2APIProxyTestContext([]byte(`{"model":"grok-4"}`))

	result, err := client.ProxyRequest(context.Background(), c, http.MethodPost, "/v1/responses", []byte(`{"model":"grok-4"}`))
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Stream)
	require.Equal(t, 11, result.Usage.InputTokens)
	require.Equal(t, 7, result.Usage.OutputTokens)
	require.Equal(t, 3, result.Usage.CacheReadInputTokens)
	require.Contains(t, rec.Body.String(), `response.completed`)
}

func TestAny2APIClientProxyRequest_ParsesChatUsageFromFinalChunk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, strings.Join([]string{
			`data: {"id":"chatcmpl_1","object":"chat.completion.chunk","model":"grok-4","choices":[{"index":0,"delta":{"content":"hi"},"finish_reason":null}]}`,
			"",
			`data: {"id":"chatcmpl_1","object":"chat.completion.chunk","model":"grok-4","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":13,"completion_tokens":5,"total_tokens":18}}`,
			"",
			"data: [DONE]",
			"",
		}, "\n"))
	}))
	defer server.Close()

	client := &Any2APIClient{enabled: true, baseURL: server.URL, apiKey: "test-key"}
	c, rec := newAny2APIProxyTestContext([]byte(`{"model":"grok-4","stream":true}`))

	result, err := client.ProxyRequest(context.Background(), c, http.MethodPost, "/v1/chat/completions", []byte(`{"model":"grok-4","stream":true}`))
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Stream)
	require.Equal(t, 13, result.Usage.InputTokens)
	require.Equal(t, 5, result.Usage.OutputTokens)
	require.Zero(t, result.Usage.CacheReadInputTokens)
	require.Contains(t, rec.Body.String(), `"usage":{"prompt_tokens":13,"completion_tokens":5,"total_tokens":18}`)
}
