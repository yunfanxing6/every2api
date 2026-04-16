package middleware

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/gemini"
	"github.com/Wei-Shaw/sub2api/internal/pkg/grok"
	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type apiKeyGroupModelLookupFunc func(ctx context.Context, groupID int64, platform string) []string

type requestedRouteInfo struct {
	path     string
	model    string
	platform string
}

var (
	apiKeyGroupModelLookupMu sync.RWMutex
	apiKeyGroupModelLookup   apiKeyGroupModelLookupFunc
)

func ConfigureAPIKeyGroupModelLookup(gatewayService *service.GatewayService) {
	if gatewayService == nil {
		setAPIKeyGroupModelLookup(nil)
		return
	}
	setAPIKeyGroupModelLookup(func(ctx context.Context, groupID int64, platform string) []string {
		gid := groupID
		return gatewayService.GetAvailableModels(ctx, &gid, platform)
	})
}

func setAPIKeyGroupModelLookup(fn apiKeyGroupModelLookupFunc) {
	apiKeyGroupModelLookupMu.Lock()
	defer apiKeyGroupModelLookupMu.Unlock()
	apiKeyGroupModelLookup = fn
}

func getAPIKeyGroupModelLookup() apiKeyGroupModelLookupFunc {
	apiKeyGroupModelLookupMu.RLock()
	defer apiKeyGroupModelLookupMu.RUnlock()
	return apiKeyGroupModelLookup
}

// NewAPIKeyAuthMiddleware 创建 API Key 认证中间件
func NewAPIKeyAuthMiddleware(apiKeyService *service.APIKeyService, subscriptionService *service.SubscriptionService, cfg *config.Config) APIKeyAuthMiddleware {
	return APIKeyAuthMiddleware(apiKeyAuthWithSubscription(apiKeyService, subscriptionService, cfg))
}

// apiKeyAuthWithSubscription API Key认证中间件（支持订阅验证）
//
// 中间件职责分为两层：
//   - 鉴权（Authentication）：验证 Key 有效性、用户状态、IP 限制 —— 始终执行
//   - 计费执行（Billing Enforcement）：过期/配额/订阅/余额检查 —— skipBilling 时整块跳过
//
// /v1/usage 端点只需鉴权，不需要计费执行（允许过期/配额耗尽的 Key 查询自身用量）。
func apiKeyAuthWithSubscription(apiKeyService *service.APIKeyService, subscriptionService *service.SubscriptionService, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ── 1. 提取 API Key ──────────────────────────────────────────

		queryKey := strings.TrimSpace(c.Query("key"))
		queryApiKey := strings.TrimSpace(c.Query("api_key"))
		if queryKey != "" || queryApiKey != "" {
			AbortWithError(c, 400, "api_key_in_query_deprecated", "API key in query parameter is deprecated. Please use Authorization header instead.")
			return
		}

		// 尝试从Authorization header中提取API key (Bearer scheme)
		authHeader := c.GetHeader("Authorization")
		var apiKeyString string

		if authHeader != "" {
			// 验证Bearer scheme
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				apiKeyString = strings.TrimSpace(parts[1])
			}
		}

		// 如果Authorization header中没有，尝试从x-api-key header中提取
		if apiKeyString == "" {
			apiKeyString = c.GetHeader("x-api-key")
		}

		// 如果x-api-key header中没有，尝试从x-goog-api-key header中提取（Gemini CLI兼容）
		if apiKeyString == "" {
			apiKeyString = c.GetHeader("x-goog-api-key")
		}

		// 如果所有header都没有API key
		if apiKeyString == "" {
			AbortWithError(c, 401, "API_KEY_REQUIRED", "API key is required in Authorization header (Bearer scheme), x-api-key header, or x-goog-api-key header")
			return
		}

		// ── 2. 验证 Key 存在 ─────────────────────────────────────────

		apiKey, err := apiKeyService.GetByKey(c.Request.Context(), apiKeyString)
		if err != nil {
			if errors.Is(err, service.ErrAPIKeyNotFound) {
				AbortWithError(c, 401, "INVALID_API_KEY", "Invalid API key")
				return
			}
			AbortWithError(c, 500, "INTERNAL_ERROR", "Failed to validate API key")
			return
		}

		// ── 3. 基础鉴权（始终执行） ─────────────────────────────────

		// disabled / 未知状态 → 无条件拦截（expired 和 quota_exhausted 留给计费阶段）
		if !apiKey.IsActive() &&
			apiKey.Status != service.StatusAPIKeyExpired &&
			apiKey.Status != service.StatusAPIKeyQuotaExhausted {
			AbortWithError(c, 401, "API_KEY_DISABLED", "API key is disabled")
			return
		}

		// 检查 IP 限制（白名单/黑名单）
		// 注意：错误信息故意模糊，避免暴露具体的 IP 限制机制
		if len(apiKey.IPWhitelist) > 0 || len(apiKey.IPBlacklist) > 0 {
			clientIP := ip.GetTrustedClientIP(c)
			allowed, _ := ip.CheckIPRestrictionWithCompiledRules(clientIP, apiKey.CompiledIPWhitelist, apiKey.CompiledIPBlacklist)
			if !allowed {
				AbortWithError(c, 403, "ACCESS_DENIED", "Access denied")
				return
			}
		}

		// 检查关联的用户
		if apiKey.User == nil {
			AbortWithError(c, 401, "USER_NOT_FOUND", "User associated with API key not found")
			return
		}

		// 检查用户状态
		if !apiKey.User.IsActive() {
			AbortWithError(c, 401, "USER_INACTIVE", "User account is not active")
			return
		}

		apiKey = resolveAPIKeyForRequest(c, apiKey)

		// ── 4. SimpleMode → early return ─────────────────────────────

		if cfg.RunMode == config.RunModeSimple {
			c.Set(string(ContextKeyAPIKey), apiKey)
			c.Set(string(ContextKeyUser), AuthSubject{
				UserID:      apiKey.User.ID,
				Concurrency: apiKey.User.Concurrency,
			})
			c.Set(string(ContextKeyUserRole), apiKey.User.Role)
			setGroupContext(c, apiKey.Group)
			_ = apiKeyService.TouchLastUsed(c.Request.Context(), apiKey.ID)
			c.Next()
			return
		}

		// ── 5. 加载订阅（订阅模式时始终加载） ───────────────────────

		// skipBilling: /v1/usage 只需鉴权，跳过所有计费执行
		skipBilling := c.Request.URL.Path == "/v1/usage"

		var subscription *service.UserSubscription
		isSubscriptionType := apiKey.Group != nil && apiKey.Group.IsSubscriptionType()

		if isSubscriptionType && subscriptionService != nil {
			sub, subErr := subscriptionService.GetActiveSubscription(
				c.Request.Context(),
				apiKey.User.ID,
				apiKey.Group.ID,
			)
			if subErr != nil {
				if !skipBilling {
					AbortWithError(c, 403, "SUBSCRIPTION_NOT_FOUND", "No active subscription found for this group")
					return
				}
				// skipBilling: 订阅不存在也放行，handler 会返回可用的数据
			} else {
				subscription = sub
			}
		}

		// ── 6. 计费执行（skipBilling 时整块跳过） ────────────────────

		if !skipBilling {
			// Key 状态检查
			switch apiKey.Status {
			case service.StatusAPIKeyQuotaExhausted:
				AbortWithError(c, 429, "API_KEY_QUOTA_EXHAUSTED", "API key 额度已用完")
				return
			case service.StatusAPIKeyExpired:
				AbortWithError(c, 403, "API_KEY_EXPIRED", "API key 已过期")
				return
			}

			// 运行时过期/配额检查（即使状态是 active，也要检查时间和用量）
			if apiKey.IsExpired() {
				AbortWithError(c, 403, "API_KEY_EXPIRED", "API key 已过期")
				return
			}
			if apiKey.IsQuotaExhausted() {
				AbortWithError(c, 429, "API_KEY_QUOTA_EXHAUSTED", "API key 额度已用完")
				return
			}

			// 订阅模式：验证订阅限额
			if subscription != nil {
				needsMaintenance, validateErr := subscriptionService.ValidateAndCheckLimits(subscription, apiKey.Group)
				if validateErr != nil {
					code := "SUBSCRIPTION_INVALID"
					status := 403
					if errors.Is(validateErr, service.ErrDailyLimitExceeded) ||
						errors.Is(validateErr, service.ErrWeeklyLimitExceeded) ||
						errors.Is(validateErr, service.ErrMonthlyLimitExceeded) {
						code = "USAGE_LIMIT_EXCEEDED"
						status = 429
					}
					AbortWithError(c, status, code, validateErr.Error())
					return
				}

				// 窗口维护异步化（不阻塞请求）
				if needsMaintenance {
					maintenanceCopy := *subscription
					subscriptionService.DoWindowMaintenance(&maintenanceCopy)
				}
			} else {
				// 非订阅模式 或 订阅模式但 subscriptionService 未注入：回退到余额检查
				if apiKey.User.Balance <= 0 {
					AbortWithError(c, 403, "INSUFFICIENT_BALANCE", "Insufficient account balance")
					return
				}
			}
		}

		// ── 7. 设置上下文 → Next ─────────────────────────────────────

		if subscription != nil {
			c.Set(string(ContextKeySubscription), subscription)
		}
		c.Set(string(ContextKeyAPIKey), apiKey)
		c.Set(string(ContextKeyUser), AuthSubject{
			UserID:      apiKey.User.ID,
			Concurrency: apiKey.User.Concurrency,
		})
		c.Set(string(ContextKeyUserRole), apiKey.User.Role)
		setGroupContext(c, apiKey.Group)
		_ = apiKeyService.TouchLastUsed(c.Request.Context(), apiKey.ID)

		c.Next()
	}
}

// GetAPIKeyFromContext 从上下文中获取API key
func GetAPIKeyFromContext(c *gin.Context) (*service.APIKey, bool) {
	value, exists := c.Get(string(ContextKeyAPIKey))
	if !exists {
		return nil, false
	}
	apiKey, ok := value.(*service.APIKey)
	return apiKey, ok
}

// GetSubscriptionFromContext 从上下文中获取订阅信息
func GetSubscriptionFromContext(c *gin.Context) (*service.UserSubscription, bool) {
	value, exists := c.Get(string(ContextKeySubscription))
	if !exists {
		return nil, false
	}
	subscription, ok := value.(*service.UserSubscription)
	return subscription, ok
}

func setGroupContext(c *gin.Context, group *service.Group) {
	if !service.IsGroupContextValid(group) {
		return
	}
	if existing, ok := c.Request.Context().Value(ctxkey.Group).(*service.Group); ok && existing != nil && existing.ID == group.ID && service.IsGroupContextValid(existing) {
		return
	}
	ctx := context.WithValue(c.Request.Context(), ctxkey.Group, group)
	c.Request = c.Request.WithContext(ctx)
}

func resolveAPIKeyForRequest(c *gin.Context, apiKey *service.APIKey) *service.APIKey {
	if apiKey == nil {
		return nil
	}
	groups := candidateAPIKeyGroups(apiKey)
	if len(groups) == 0 {
		return apiKey
	}
	if len(groups) == 1 {
		if apiKey.Group == nil || apiKey.Group.ID != groups[0].ID {
			return cloneAPIKeyWithResolvedGroup(apiKey, groups[0])
		}
		return apiKey
	}

	info := resolveRequestedRouteInfo(c)
	candidates := filterCandidateGroupsByPlatform(groups, info.platform)
	if len(candidates) == 0 {
		return apiKey
	}
	selected := selectBestGroupForRequest(c.Request.Context(), apiKey, candidates, info)
	if selected == nil {
		selected = preferPrimaryCandidate(apiKey, candidates)
	}
	if selected == nil {
		return apiKey
	}
	if apiKey.Group != nil && apiKey.Group.ID == selected.ID && service.IsGroupContextValid(apiKey.Group) {
		return apiKey
	}
	return cloneAPIKeyWithResolvedGroup(apiKey, selected)
}

func candidateAPIKeyGroups(apiKey *service.APIKey) []*service.Group {
	if apiKey == nil {
		return nil
	}
	if len(apiKey.Groups) > 0 {
		groups := make([]*service.Group, 0, len(apiKey.Groups))
		for i := range apiKey.Groups {
			groups = append(groups, &apiKey.Groups[i])
		}
		return groups
	}
	if apiKey.Group != nil {
		return []*service.Group{apiKey.Group}
	}
	return nil
}

func cloneAPIKeyWithResolvedGroup(apiKey *service.APIKey, group *service.Group) *service.APIKey {
	if apiKey == nil || group == nil {
		return apiKey
	}
	cloned := *apiKey
	groupID := group.ID
	cloned.GroupID = &groupID
	cloned.Group = group
	if len(cloned.GroupIDs) == 0 {
		cloned.GroupIDs = apiKey.EffectiveGroupIDs()
	}
	return &cloned
}

func resolveRequestedRouteInfo(c *gin.Context) requestedRouteInfo {
	info := requestedRouteInfo{}
	if c == nil {
		return info
	}
	if forcePlatform, ok := GetForcePlatformFromContext(c); ok && strings.TrimSpace(forcePlatform) != "" {
		info.platform = strings.TrimSpace(forcePlatform)
	}
	info.path = c.FullPath()
	if info.path == "" && c.Request != nil && c.Request.URL != nil {
		info.path = c.Request.URL.Path
	}
	info.model = strings.TrimSpace(extractRequestedModel(c, info.path))
	if info.platform == "" {
		info.platform = resolveRequestedPlatformFromInfo(info)
	}
	return info
}

func resolveRequestedPlatform(c *gin.Context) string {
	return resolveRequestedPlatformFromInfo(resolveRequestedRouteInfo(c))
}

func resolveRequestedPlatformFromInfo(info requestedRouteInfo) string {
	if strings.TrimSpace(info.platform) != "" {
		return strings.TrimSpace(info.platform)
	}
	path := info.path
	if strings.HasPrefix(path, "/antigravity/") {
		return service.PlatformAntigravity
	}
	if strings.HasPrefix(path, "/v1beta/") {
		return service.PlatformGemini
	}
	if info.model != "" {
		if platform := inferPlatformFromModel(info.model); platform != "" {
			return platform
		}
	}
	switch {
	case strings.HasSuffix(path, "/messages") || strings.HasSuffix(path, "/messages/count_tokens"):
		return service.PlatformAnthropic
	case strings.HasSuffix(path, "/chat/completions"),
		strings.HasSuffix(path, "/responses"),
		strings.Contains(path, "/responses/"),
		strings.HasSuffix(path, "/images/generations"),
		strings.HasSuffix(path, "/images/edits"),
		strings.HasSuffix(path, "/videos"),
		strings.HasSuffix(path, "/video/extend"):
		return service.PlatformOpenAI
	default:
		return ""
	}
}

func filterCandidateGroupsByPlatform(groups []*service.Group, desiredPlatform string) []*service.Group {
	if desiredPlatform == "" {
		return groups
	}
	filtered := make([]*service.Group, 0, len(groups))
	for _, group := range groups {
		if group == nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(group.Platform), desiredPlatform) {
			filtered = append(filtered, group)
		}
	}
	return filtered
}

func selectBestGroupForRequest(ctx context.Context, apiKey *service.APIKey, groups []*service.Group, info requestedRouteInfo) *service.Group {
	if len(groups) == 0 {
		return nil
	}
	primaryID := int64(0)
	if apiKey != nil && apiKey.GroupID != nil {
		primaryID = *apiKey.GroupID
	}
	groupOrder := apiKeyGroupOrder(apiKey)
	var selected *service.Group
	bestScore := -1
	bestRank := len(groupOrder) + 1
	for _, group := range groups {
		score := scoreGroupForRequest(ctx, group, info, primaryID)
		if score < 0 {
			continue
		}
		rank := groupOrder[group.ID]
		if selected == nil || score > bestScore || (score == bestScore && rank < bestRank) {
			selected = group
			bestScore = score
			bestRank = rank
		}
	}
	return selected
}

func apiKeyGroupOrder(apiKey *service.APIKey) map[int64]int {
	order := make(map[int64]int)
	if apiKey == nil {
		return order
	}
	for idx, id := range apiKey.EffectiveGroupIDs() {
		order[id] = idx
	}
	return order
}

func preferPrimaryCandidate(apiKey *service.APIKey, candidates []*service.Group) *service.Group {
	if apiKey != nil && apiKey.GroupID != nil {
		for _, group := range candidates {
			if group != nil && group.ID == *apiKey.GroupID {
				return group
			}
		}
	}
	if len(candidates) > 0 {
		return candidates[0]
	}
	return nil
}

func scoreGroupForRequest(ctx context.Context, group *service.Group, info requestedRouteInfo, primaryGroupID int64) int {
	if group == nil || !group.IsActive() {
		return -1
	}
	score := 0
	if primaryGroupID > 0 && group.ID == primaryGroupID {
		score += 5
	}
	if isMessagesEndpoint(info.path) && isOpenAICompatibleGroupPlatform(group.Platform) {
		if !group.AllowMessagesDispatch {
			return -1
		}
		score += 20
	}
	if info.model == "" {
		return score + 1
	}
	if explicitModels := getAvailableModelsForGroup(ctx, group); len(explicitModels) > 0 {
		if requestedModelMatchesAny(group.Platform, info.model, explicitModels) {
			return score + 100
		}
		if modelMatchesPlatform(group.Platform, info.model, group.DefaultMappedModel) {
			return score + 85
		}
		return -1
	}
	if len(group.GetRoutingAccountIDs(info.model)) > 0 {
		score += 90
	}
	if modelMatchesPlatform(group.Platform, info.model, group.DefaultMappedModel) {
		score += 80
	}
	if supportsRequestedModelByDefault(group, info) {
		score += 70
	}
	if score == 0 && info.platform != "" && strings.EqualFold(strings.TrimSpace(group.Platform), info.platform) {
		score = 10
	}
	return score
}

func getAvailableModelsForGroup(ctx context.Context, group *service.Group) []string {
	if group == nil {
		return nil
	}
	lookup := getAPIKeyGroupModelLookup()
	if lookup == nil {
		return nil
	}
	return lookup(ctx, group.ID, strings.TrimSpace(group.Platform))
}

func requestedModelMatchesAny(platform, requestedModel string, candidates []string) bool {
	for _, candidate := range candidates {
		if modelMatchesPlatform(platform, requestedModel, candidate) {
			return true
		}
	}
	return false
}

func modelMatchesPlatform(platform, left, right string) bool {
	left = normalizePlatformModel(platform, left)
	right = normalizePlatformModel(platform, right)
	return left != "" && left == right
}

func normalizePlatformModel(platform, model string) string {
	trimmed := strings.TrimSpace(model)
	if trimmed == "" {
		return ""
	}
	trimmed = strings.TrimPrefix(trimmed, "models/")
	switch strings.TrimSpace(platform) {
	case service.PlatformAnthropic:
		return claude.NormalizeModelID(trimmed)
	default:
		return trimmed
	}
}

func supportsRequestedModelByDefault(group *service.Group, info requestedRouteInfo) bool {
	if group == nil {
		return false
	}
	platform := strings.TrimSpace(group.Platform)
	switch platform {
	case service.PlatformAnthropic:
		return requestedModelMatchesAny(platform, info.model, claude.DefaultModelIDs())
	case service.PlatformGemini:
		return gemini.HasFallbackModel(info.model)
	case service.PlatformOpenAI:
		return requestedModelMatchesAny(platform, info.model, openai.DefaultModelIDs())
	case service.PlatformGrok:
		return requestedModelMatchesAny(platform, info.model, grok.DefaultModelIDs())
	case service.PlatformQwen:
		return strings.HasPrefix(strings.ToLower(strings.TrimSpace(info.model)), "qwen")
	case service.PlatformAntigravity:
		return supportsAntigravityScope(group, info)
	default:
		return false
	}
}

func supportsAntigravityScope(group *service.Group, info requestedRouteInfo) bool {
	if group == nil {
		return false
	}
	if len(group.SupportedModelScopes) == 0 {
		return true
	}
	scopes := make([]string, 0, len(group.SupportedModelScopes))
	for _, scope := range group.SupportedModelScopes {
		scopes = append(scopes, strings.TrimSpace(scope))
	}
	loweredModel := strings.ToLower(strings.TrimSpace(info.model))
	if strings.HasPrefix(loweredModel, "claude") {
		return slices.Contains(scopes, "claude")
	}
	if strings.HasPrefix(loweredModel, "gemini") {
		if strings.Contains(loweredModel, "image") {
			return slices.Contains(scopes, "gemini_image")
		}
		return slices.Contains(scopes, "gemini_text") || slices.Contains(scopes, "gemini_image")
	}
	if strings.HasPrefix(info.path, "/antigravity/v1beta") {
		return slices.Contains(scopes, "gemini_text") || slices.Contains(scopes, "gemini_image")
	}
	return slices.Contains(scopes, "claude")
}

func isMessagesEndpoint(path string) bool {
	return strings.HasSuffix(path, "/messages") || strings.HasSuffix(path, "/messages/count_tokens")
}

func isOpenAICompatibleGroupPlatform(platform string) bool {
	return platform == service.PlatformOpenAI || platform == service.PlatformGrok || platform == service.PlatformQwen
}

func extractRequestedModel(c *gin.Context, path string) string {
	if c == nil {
		return ""
	}
	if strings.HasPrefix(path, "/v1beta/models/") {
		if modelAction := strings.TrimSpace(strings.TrimPrefix(c.Param("modelAction"), "/")); modelAction != "" {
			if idx := strings.IndexByte(modelAction, ':'); idx > 0 {
				return strings.TrimSpace(modelAction[:idx])
			}
		}
		return strings.TrimSpace(strings.TrimPrefix(c.Param("model"), "/"))
	}

	if c.Request == nil || c.Request.Body == nil || c.Request.Method == http.MethodGet {
		return ""
	}
	body, err := pkghttputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil || len(body) == 0 {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	c.Request.ContentLength = int64(len(body))

	protocol := ""
	if strings.HasPrefix(path, "/v1beta/") {
		protocol = service.PlatformGemini
	}
	parsed, err := service.ParseGatewayRequest(body, protocol)
	if err != nil || parsed == nil {
		return ""
	}
	return strings.TrimSpace(parsed.Model)
}

func inferPlatformFromModel(model string) string {
	lowered := strings.ToLower(strings.TrimSpace(model))
	switch {
	case lowered == "":
		return ""
	case strings.HasPrefix(lowered, "claude"):
		return service.PlatformAnthropic
	case strings.HasPrefix(lowered, "gemini"):
		return service.PlatformGemini
	case strings.HasPrefix(lowered, "grok"):
		return service.PlatformGrok
	case strings.HasPrefix(lowered, "qwen"):
		return service.PlatformQwen
	case strings.HasPrefix(lowered, "gpt"), strings.HasPrefix(lowered, "o1"), strings.HasPrefix(lowered, "o3"), strings.HasPrefix(lowered, "o4"):
		return service.PlatformOpenAI
	default:
		return ""
	}
}
