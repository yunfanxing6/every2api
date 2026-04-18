package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

type Any2APIModel struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	OwnedBy     string `json:"owned_by"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type Any2APIProxyResult struct {
	Usage      OpenAIUsage
	Stream     bool
	MediaType  string
	ImageCount int
	ImageSize  string
	Duration   time.Duration
}

type Any2APITestResult struct {
	Model string
	Text  string
}

type Any2APISummary struct {
	Enabled            bool           `json:"enabled"`
	Connected          bool           `json:"connected"`
	BaseURL            string         `json:"base_url"`
	ServiceName        string         `json:"service_name,omitempty"`
	Version            string         `json:"version,omitempty"`
	ModelsTotal        int            `json:"models_total"`
	AccountsTotal      int            `json:"accounts_total"`
	ManageableAccounts int            `json:"manageable_accounts"`
	SelectableAccounts int            `json:"selectable_accounts"`
	Providers          map[string]int `json:"providers,omitempty"`
	Pools              map[string]int `json:"pools,omitempty"`
	Statuses           map[string]int `json:"statuses,omitempty"`
	Error              string         `json:"error,omitempty"`
}

func (s Any2APISummary) ProviderConnected(provider string) bool {
	if !s.Enabled || !s.Connected {
		return false
	}
	return s.Providers[strings.ToLower(strings.TrimSpace(provider))] > 0
}

type Any2APIClient struct {
	enabled bool
	baseURL string
	apiKey  string
	client  *http.Client
}

type any2apiSummaryEnvelope struct {
	Status  string `json:"status"`
	Service struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"service"`
	Models struct {
		Total     int            `json:"total"`
		Providers map[string]int `json:"providers"`
	} `json:"models"`
	Accounts struct {
		Revision   int            `json:"revision"`
		Total      int            `json:"total"`
		Manageable int            `json:"manageable"`
		Selectable int            `json:"selectable"`
		Pools      map[string]int `json:"pools"`
		Statuses   map[string]int `json:"statuses"`
	} `json:"accounts"`
}

type any2apiModelsEnvelope struct {
	Object string         `json:"object"`
	Data   []Any2APIModel `json:"data"`
}

func NewAny2APIClientWithSecret(settings Any2APISettings, apiKey string) *Any2APIClient {
	baseURL := strings.TrimRight(strings.TrimSpace(settings.BaseURL), "/")
	timeoutSeconds := settings.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}
	return &Any2APIClient{
		enabled: settings.Enabled,
		baseURL: baseURL,
		apiKey:  strings.TrimSpace(apiKey),
		client: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}
}

func (c *Any2APIClient) Summary(ctx context.Context) Any2APISummary {
	summary := Any2APISummary{
		Enabled: c != nil && c.enabled,
		BaseURL: "",
	}
	if c == nil {
		summary.Error = "client not initialized"
		return summary
	}
	summary.BaseURL = c.baseURL
	if !c.enabled {
		return summary
	}
	if c.baseURL == "" {
		summary.Error = "any2api base URL is not configured"
		return summary
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/internal/providers/summary", nil)
	if err != nil {
		summary.Error = err.Error()
		return summary
	}
	if c.apiKey != "" {
		req.Header.Set("X-Internal-Key", c.apiKey)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "every2api-admin-sync/0.1")

	resp, err := c.client.Do(req)
	if err != nil {
		summary.Error = err.Error()
		return summary
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		summary.Error = fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		return summary
	}

	var payload any2apiSummaryEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		summary.Error = err.Error()
		return summary
	}

	summary.Connected = true
	summary.ServiceName = payload.Service.Name
	summary.Version = payload.Service.Version
	summary.ModelsTotal = payload.Models.Total
	summary.AccountsTotal = payload.Accounts.Total
	summary.ManageableAccounts = payload.Accounts.Manageable
	summary.SelectableAccounts = payload.Accounts.Selectable
	summary.Providers = payload.Models.Providers
	summary.Pools = payload.Accounts.Pools
	summary.Statuses = payload.Accounts.Statuses
	return summary
}

func (c *Any2APIClient) Enabled() bool {
	return c != nil && c.enabled && c.baseURL != "" && c.apiKey != ""
}

func (c *Any2APIClient) HandlesModel(model string) bool {
	if !c.Enabled() {
		return false
	}
	lowered := strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(lowered, "qwen") || strings.HasPrefix(lowered, "grok")
}

func (c *Any2APIClient) ListModels(ctx context.Context) ([]Any2APIModel, error) {
	if !c.Enabled() {
		return nil, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "every2api-model-sync/0.1")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var payload any2apiModelsEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload.Data, nil
}

func extractAny2APIUsageFromJSON(body []byte) (OpenAIUsage, bool) {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return OpenAIUsage{}, false
	}
	if usage := gjson.GetBytes(body, "usage"); usage.Exists() {
		input := int(usage.Get("input_tokens").Int())
		if input == 0 {
			input = int(usage.Get("prompt_tokens").Int())
		}
		output := int(usage.Get("output_tokens").Int())
		if output == 0 {
			output = int(usage.Get("completion_tokens").Int())
		}
		cached := int(usage.Get("input_tokens_details.cached_tokens").Int())
		if cached == 0 {
			cached = int(usage.Get("cache_read_input_tokens").Int())
		}
		created := int(usage.Get("cache_creation_input_tokens").Int())
		return OpenAIUsage{
			InputTokens:              input,
			OutputTokens:             output,
			CacheReadInputTokens:     cached,
			CacheCreationInputTokens: created,
		}, true
	}
	if usage := gjson.GetBytes(body, "response.usage"); usage.Exists() {
		return OpenAIUsage{
			InputTokens:          int(usage.Get("input_tokens").Int()),
			OutputTokens:         int(usage.Get("output_tokens").Int()),
			CacheReadInputTokens: int(usage.Get("input_tokens_details.cached_tokens").Int()),
		}, true
	}
	return OpenAIUsage{}, false
}

func updateAny2APIUsageFromSSEPayload(payload []byte, usage *OpenAIUsage) {
	if usage == nil || len(payload) == 0 || bytes.Equal(payload, []byte("[DONE]")) {
		return
	}
	if parsed, ok := extractAny2APIUsageFromJSON(payload); ok {
		*usage = parsed
	}
}

func updateAny2APIUsageFromSSEEvent(event []byte, usage *OpenAIUsage) {
	if usage == nil || len(event) == 0 {
		return
	}
	lines := bytes.Split(event, []byte("\n"))
	payload := bytes.Buffer{}
	for _, line := range lines {
		line = bytes.TrimSuffix(line, []byte("\r"))
		data, ok := extractOpenAISSEDataLine(string(line))
		if !ok {
			continue
		}
		if payload.Len() > 0 {
			payload.WriteByte('\n')
		}
		payload.WriteString(data)
	}
	updateAny2APIUsageFromSSEPayload(payload.Bytes(), usage)
}

func (c *Any2APIClient) ProxyRequest(ctx context.Context, ginCtx *gin.Context, method, path string, body []byte) (*Any2APIProxyResult, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("any2api integration is not enabled")
	}
	startedAt := time.Now()
	result := &Any2APIProxyResult{}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	for _, header := range []string{"Content-Type", "Accept", "OpenAI-Beta", "User-Agent"} {
		if value := strings.TrimSpace(ginCtx.GetHeader(header)); value != "" {
			req.Header.Set(header, value)
		}
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	for _, header := range []string{"Content-Type", "Cache-Control", "Content-Length", "X-Accel-Buffering"} {
		for _, value := range resp.Header.Values(header) {
			ginCtx.Writer.Header().Add(header, value)
		}
	}
	ginCtx.Status(resp.StatusCode)

	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "text/event-stream") {
		result.Stream = true
		buf := make([]byte, 32*1024)
		sseBuf := make([]byte, 0, 64*1024)
		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				sseBuf = append(sseBuf, buf[:n]...)
				for {
					idx := bytes.Index(sseBuf, []byte("\n\n"))
					if idx < 0 {
						break
					}
					event := bytes.TrimSpace(sseBuf[:idx])
					sseBuf = sseBuf[idx+2:]
					updateAny2APIUsageFromSSEEvent(event, &result.Usage)
				}
				if _, writeErr := ginCtx.Writer.Write(buf[:n]); writeErr != nil {
					return nil, writeErr
				}
				ginCtx.Writer.Flush()
			}
			if readErr != nil {
				if readErr == io.EOF {
					break
				}
				return nil, readErr
			}
		}
		result.Duration = time.Since(startedAt)
		return result, nil
	}

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if usage, ok := extractAny2APIUsageFromJSON(payload); ok {
		result.Usage = usage
	}
	if strings.Contains(path, "/images/generations") {
		result.MediaType = "image"
		result.ImageCount = int(gjson.GetBytes(payload, "data.#").Int())
		result.ImageSize = gjson.GetBytes(body, "size").String()
		if result.ImageCount == 0 {
			result.ImageCount = int(gjson.GetBytes(body, "n").Int())
			if result.ImageCount == 0 {
				result.ImageCount = 1
			}
		}
	}
	_, err = ginCtx.Writer.Write(payload)
	if err != nil {
		return nil, err
	}
	result.Duration = time.Since(startedAt)
	return result, nil
}

func (c *Any2APIClient) TestConnection(ctx context.Context, modelID string, prompt string) (*Any2APITestResult, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("any2api integration is not enabled")
	}
	if strings.TrimSpace(modelID) == "" {
		return nil, fmt.Errorf("model is required")
	}
	if strings.TrimSpace(prompt) == "" {
		prompt = "hi"
	}
	payload := map[string]any{
		"model":  modelID,
		"stream": false,
		"messages": []map[string]any{{
			"role":    "user",
			"content": prompt,
		}},
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "every2api-account-test/0.1")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	text := gjson.GetBytes(respBody, "choices.0.message.content").String()
	if text == "" {
		text = string(respBody)
	}
	return &Any2APITestResult{Model: modelID, Text: text}, nil
}
