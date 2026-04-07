package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path"
	"strconv"
	"strings"

	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

const (
	grokImageGenerationModel = "grok-imagine-1.0-fast"
	grokImageEditModel       = "grok-imagine-1.0-edit"
	grokVideoModel           = "grok-imagine-1.0-video"
)

type grokMediaRequestMeta struct {
	Model        string
	ImageSize    string
	ImageCount   int
	Stream       bool
	VideoSeconds int
	VideoQuality string
}

func normalizeGrokGatewayPlatform(apiKey *service.APIKey) string {
	if apiKey != nil && apiKey.Group != nil && strings.TrimSpace(apiKey.Group.Platform) != "" {
		return apiKey.Group.Platform
	}
	return service.PlatformOpenAI
}

func normalizeGrokImageBillingSize(size string) string {
	switch strings.TrimSpace(size) {
	case "1792x1024", "1024x1792":
		return "2K"
	default:
		return "1K"
	}
}

func parseMultipartFields(contentType string, body []byte) (map[string]string, error) {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(mediaType, "multipart/") {
		return map[string]string{}, nil
	}
	reader := multipart.NewReader(bytes.NewReader(body), params["boundary"])
	fields := map[string]string{}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if part.FileName() != "" {
			_, _ = io.Copy(io.Discard, part)
			_ = part.Close()
			continue
		}
		value, err := io.ReadAll(io.LimitReader(part, 1<<20))
		_ = part.Close()
		if err != nil {
			return nil, err
		}
		fields[part.FormName()] = strings.TrimSpace(string(value))
	}
	return fields, nil
}

func parseGrokImageRequestMeta(body []byte, defaultModel string) grokMediaRequestMeta {
	model := strings.TrimSpace(gjson.GetBytes(body, "model").String())
	if model == "" {
		model = defaultModel
	}
	size := strings.TrimSpace(gjson.GetBytes(body, "size").String())
	if size == "" {
		size = "1024x1024"
	}
	count := int(gjson.GetBytes(body, "n").Int())
	if count <= 0 {
		count = 1
	}
	return grokMediaRequestMeta{
		Model:      model,
		ImageSize:  normalizeGrokImageBillingSize(size),
		ImageCount: count,
		Stream:     gjson.GetBytes(body, "stream").Bool(),
	}
}

func parseGrokImageEditRequestMeta(body []byte, contentType string) (grokMediaRequestMeta, error) {
	fields, err := parseMultipartFields(contentType, body)
	if err != nil {
		return grokMediaRequestMeta{}, err
	}
	model := strings.TrimSpace(fields["model"])
	if model == "" {
		model = grokImageEditModel
	}
	size := strings.TrimSpace(fields["size"])
	if size == "" {
		size = "1024x1024"
	}
	count := 1
	if raw := strings.TrimSpace(fields["n"]); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			count = parsed
		}
	}
	stream := false
	if raw := strings.TrimSpace(fields["stream"]); raw != "" {
		stream = strings.EqualFold(raw, "true") || raw == "1"
	}
	return grokMediaRequestMeta{
		Model:      model,
		ImageSize:  normalizeGrokImageBillingSize(size),
		ImageCount: count,
		Stream:     stream,
	}, nil
}

func parseGrokVideoRequestMeta(body []byte, contentType string) (grokMediaRequestMeta, error) {
	trimmedContentType := strings.ToLower(strings.TrimSpace(contentType))
	if strings.HasPrefix(trimmedContentType, "application/json") {
		model := strings.TrimSpace(gjson.GetBytes(body, "model").String())
		if model == "" {
			model = grokVideoModel
		}
		seconds := int(gjson.GetBytes(body, "seconds").Int())
		if seconds <= 0 {
			seconds = 6
		}
		quality := strings.TrimSpace(gjson.GetBytes(body, "quality").String())
		if quality == "" {
			quality = "standard"
		}
		return grokMediaRequestMeta{Model: model, VideoSeconds: seconds, VideoQuality: quality}, nil
	}
	fields, err := parseMultipartFields(contentType, body)
	if err != nil {
		return grokMediaRequestMeta{}, err
	}
	model := strings.TrimSpace(fields["model"])
	if model == "" {
		model = grokVideoModel
	}
	seconds := 6
	if raw := strings.TrimSpace(fields["seconds"]); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			seconds = parsed
		}
	}
	quality := strings.TrimSpace(fields["quality"])
	if quality == "" {
		quality = "standard"
	}
	return grokMediaRequestMeta{Model: model, VideoSeconds: seconds, VideoQuality: quality}, nil
}

func rewriteGrokMediaBodyForUpstream(body []byte, contentType string, upstreamModel string) ([]byte, string, error) {
	upstreamModel = strings.TrimSpace(upstreamModel)
	if upstreamModel == "" || len(body) == 0 {
		return body, contentType, nil
	}
	trimmedContentType := strings.ToLower(strings.TrimSpace(contentType))
	if strings.HasPrefix(trimmedContentType, "application/json") || trimmedContentType == "" {
		payload := map[string]any{}
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, contentType, err
		}
		payload["model"] = upstreamModel
		rewritten, err := json.Marshal(payload)
		return rewritten, contentType, err
	}
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, contentType, err
	}
	if !strings.HasPrefix(mediaType, "multipart/") {
		return body, contentType, nil
	}
	originalReader := multipart.NewReader(bytes.NewReader(body), params["boundary"])
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	modelWritten := false
	for {
		part, err := originalReader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, contentType, err
		}
		header := make(textproto.MIMEHeader, len(part.Header))
		for key, values := range part.Header {
			copied := append([]string(nil), values...)
			header[key] = copied
		}
		newPart, err := writer.CreatePart(header)
		if err != nil {
			_ = part.Close()
			return nil, contentType, err
		}
		if part.FormName() == "model" && part.FileName() == "" {
			if _, err := io.WriteString(newPart, upstreamModel); err != nil {
				_ = part.Close()
				return nil, contentType, err
			}
			modelWritten = true
		} else if _, err := io.Copy(newPart, part); err != nil {
			_ = part.Close()
			return nil, contentType, err
		}
		_ = part.Close()
	}
	if !modelWritten {
		fieldWriter, err := writer.CreateFormField("model")
		if err != nil {
			return nil, contentType, err
		}
		if _, err := io.WriteString(fieldWriter, upstreamModel); err != nil {
			return nil, contentType, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, contentType, err
	}
	return buf.Bytes(), writer.FormDataContentType(), nil
}

func requestExternalBaseURL(c *gin.Context, settingService *service.SettingService) string {
	if settingService != nil {
		if configured := strings.TrimSpace(settingService.GetAPIBaseURL(c.Request.Context())); configured != "" {
			return strings.TrimRight(configured, "/")
		}
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if forwardedProto := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")); forwardedProto != "" {
		scheme = forwardedProto
	}
	return scheme + "://" + c.Request.Host
}

func upstreamFilesPrefix(baseURL string) string {
	normalized := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(normalized, "/v1") {
		return normalized + "/files/"
	}
	return normalized + "/v1/files/"
}

func rewriteGrokMediaBody(c *gin.Context, settingService *service.SettingService, account *service.Account, body []byte) []byte {
	baseURL := strings.TrimSpace(account.GetOpenAIBaseURL())
	if baseURL == "" || len(body) == 0 {
		return body
	}
	upstreamPrefix := upstreamFilesPrefix(baseURL)
	publicPrefix := requestExternalBaseURL(c, settingService) + "/v1/files/"
	rewritten := strings.ReplaceAll(string(body), upstreamPrefix, publicPrefix)
	rewritten = strings.ReplaceAll(rewritten, "\"/v1/files/", "\""+publicPrefix)
	return []byte(rewritten)
}

func writeCompatibleMediaResponse(c *gin.Context, result *service.OpenAIMediaProxyResult, body []byte) {
	for _, header := range []string{"Content-Type", "Cache-Control", "X-Request-Id"} {
		for _, value := range result.Headers.Values(header) {
			c.Writer.Header().Add(header, value)
		}
	}
	c.Status(result.StatusCode)
	_, _ = c.Writer.Write(body)
}

func (h *OpenAIGatewayHandler) handleGrokMediaCreate(
	c *gin.Context,
	upstreamPath string,
	parseMeta func([]byte, string) (grokMediaRequestMeta, error),
	mediaType string,
) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	if normalizeGrokGatewayPlatform(apiKey) != service.PlatformGrok {
		h.errorResponse(c, http.StatusNotFound, "not_found_error", "This endpoint is only available for Grok groups")
		return
	}
	c.Request = c.Request.WithContext(service.WithOpenAICompatiblePlatform(c.Request.Context(), service.PlatformGrok))

	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusInternalServerError, "api_error", "User context not found")
		return
	}
	reqLog := requestLogger(
		c,
		"handler.openai_gateway.grok_media",
		zap.Int64("user_id", subject.UserID),
		zap.Int64("api_key_id", apiKey.ID),
		zap.Any("group_id", apiKey.GroupID),
	)

	body, err := pkghttputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}
	if len(body) == 0 {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Request body is empty")
		return
	}
	contentType := c.GetHeader("Content-Type")
	meta, err := parseMeta(body, contentType)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
		return
	}
	if meta.Stream {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Streaming media generation is not supported by the Grok gateway yet")
		return
	}

	setOpsRequestContext(c, meta.Model, false, body)
	setOpsEndpointContext(c, "", int16(service.RequestTypeFromLegacy(false, false)))

	subscription, _ := middleware2.GetSubscriptionFromContext(c)
	userReleaseFunc, acquired := h.acquireResponsesUserSlot(c, subject.UserID, subject.Concurrency, false, new(bool), reqLog)
	if !acquired {
		return
	}
	if userReleaseFunc != nil {
		defer userReleaseFunc()
	}
	if err := h.billingCacheService.CheckBillingEligibility(c.Request.Context(), apiKey.User, apiKey, apiKey.Group, subscription); err != nil {
		status, code, message := billingErrorDetails(err)
		h.handleStreamingAwareError(c, status, code, message, false)
		return
	}

	sessionHash := h.gatewayService.GenerateSessionHash(c, body)
	failedAccountIDs := make(map[int64]struct{})
	var lastFailoverErr *service.UpstreamFailoverError

	for {
		selection, _, err := h.gatewayService.SelectAccountWithScheduler(
			c.Request.Context(),
			apiKey.GroupID,
			"",
			sessionHash,
			meta.Model,
			failedAccountIDs,
			service.OpenAIUpstreamTransportAny,
		)
		if err != nil {
			if lastFailoverErr != nil {
				h.handleFailoverExhausted(c, lastFailoverErr, false)
			} else {
				h.handleStreamingAwareError(c, http.StatusServiceUnavailable, "api_error", "No available accounts", false)
			}
			return
		}
		account := selection.Account
		if account == nil {
			h.handleStreamingAwareError(c, http.StatusServiceUnavailable, "api_error", "No available accounts", false)
			return
		}
		upstreamModel := meta.Model
		if mapped, matched := account.ResolveMappedModel(meta.Model); matched {
			upstreamModel = mapped
		}
		forwardBody, forwardContentType, err := rewriteGrokMediaBodyForUpstream(body, contentType, upstreamModel)
		if err != nil {
			h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to normalize upstream media request")
			return
		}
		accountReleaseFunc, acquired := h.acquireResponsesAccountSlot(c, apiKey.GroupID, sessionHash, selection, false, new(bool), reqLog)
		if !acquired {
			return
		}
		proxyResult, err := h.gatewayService.ForwardCompatibleMedia(c.Request.Context(), c, account, http.MethodPost, upstreamPath, forwardContentType, forwardBody)
		if accountReleaseFunc != nil {
			accountReleaseFunc()
		}
		if err != nil {
			var failoverErr *service.UpstreamFailoverError
			if errors.As(err, &failoverErr) {
				failedAccountIDs[account.ID] = struct{}{}
				lastFailoverErr = failoverErr
				continue
			}
			h.handleStreamingAwareError(c, http.StatusBadGateway, "upstream_error", "Upstream request failed", false)
			return
		}

		responseBody := rewriteGrokMediaBody(c, h.settingService, account, proxyResult.Body)
		writeCompatibleMediaResponse(c, proxyResult, responseBody)
		if proxyResult.StatusCode >= 400 {
			return
		}

		if dataCount := int(gjson.GetBytes(responseBody, "data.#").Int()); dataCount > 0 {
			meta.ImageCount = dataCount
		}
		mediaURL := strings.TrimSpace(gjson.GetBytes(responseBody, "url").String())
		if mediaURL == "" {
			mediaURL = strings.TrimSpace(gjson.GetBytes(responseBody, "data.0.url").String())
		}
		result := &service.OpenAIForwardResult{
			RequestID:     proxyResult.RequestID,
			Model:         meta.Model,
			UpstreamModel: upstreamModel,
			Duration:      proxyResult.Duration,
			MediaType:     mediaType,
			ImageCount:    meta.ImageCount,
			ImageSize:     meta.ImageSize,
			MediaURL:      mediaURL,
			VideoSeconds:  meta.VideoSeconds,
			VideoQuality:  meta.VideoQuality,
		}
		h.submitUsageRecordTask(func(ctx context.Context) {
			_ = h.gatewayService.RecordUsage(ctx, &service.OpenAIRecordUsageInput{
				Result:           result,
				APIKey:           apiKey,
				User:             apiKey.User,
				Account:          account,
				Subscription:     subscription,
				InboundEndpoint:  c.FullPath(),
				UpstreamEndpoint: upstreamPath,
				UserAgent:        c.GetHeader("User-Agent"),
				IPAddress:        ip.GetClientIP(c),
				APIKeyService:    h.apiKeyService,
			})
		})
		return
	}
}

func (h *OpenAIGatewayHandler) ImageGenerations(c *gin.Context) {
	h.handleGrokMediaCreate(c, "/images/generations", func(body []byte, _ string) (grokMediaRequestMeta, error) {
		if !gjson.ValidBytes(body) {
			return grokMediaRequestMeta{}, io.ErrUnexpectedEOF
		}
		return parseGrokImageRequestMeta(body, grokImageGenerationModel), nil
	}, "image")
}

func (h *OpenAIGatewayHandler) ImageEdits(c *gin.Context) {
	h.handleGrokMediaCreate(c, "/images/edits", parseGrokImageEditRequestMeta, "image")
}

func (h *OpenAIGatewayHandler) Videos(c *gin.Context) {
	h.handleGrokMediaCreate(c, "/videos", parseGrokVideoRequestMeta, "video")
}

func (h *OpenAIGatewayHandler) VideoExtend(c *gin.Context) {
	h.handleGrokMediaCreate(c, "/video/extend", parseGrokVideoRequestMeta, "video")
}

func (h *OpenAIGatewayHandler) ProxyGrokFile(c *gin.Context, mediaType string) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	if normalizeGrokGatewayPlatform(apiKey) != service.PlatformGrok {
		h.errorResponse(c, http.StatusNotFound, "not_found_error", "This endpoint is only available for Grok groups")
		return
	}
	c.Request = c.Request.WithContext(service.WithOpenAICompatiblePlatform(c.Request.Context(), service.PlatformGrok))
	selection, _, err := h.gatewayService.SelectAccountWithScheduler(
		c.Request.Context(),
		apiKey.GroupID,
		"",
		"",
		"",
		nil,
		service.OpenAIUpstreamTransportAny,
	)
	if err != nil || selection == nil || selection.Account == nil {
		h.errorResponse(c, http.StatusServiceUnavailable, "api_error", "No available accounts")
		return
	}
	rawFilename := strings.TrimSpace(strings.TrimPrefix(c.Param("filename"), "/"))
	if rawFilename == "" || strings.Contains(rawFilename, "\x00") {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Invalid media path")
		return
	}
	cleanFilename := strings.TrimPrefix(path.Clean("/"+rawFilename), "/")
	if cleanFilename == "" || cleanFilename == "." || cleanFilename != rawFilename {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Invalid media path")
		return
	}
	upstreamPath := "/files/" + mediaType + "/" + cleanFilename
	resp, err := h.gatewayService.ProxyCompatibleMediaFile(c.Request.Context(), selection.Account, upstreamPath)
	if err != nil {
		h.handleStreamingAwareError(c, http.StatusBadGateway, "upstream_error", "Failed to fetch upstream media", false)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	for _, header := range []string{"Content-Type", "Cache-Control", "Content-Length"} {
		for _, value := range resp.Header.Values(header) {
			c.Writer.Header().Add(header, value)
		}
	}
	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
}
