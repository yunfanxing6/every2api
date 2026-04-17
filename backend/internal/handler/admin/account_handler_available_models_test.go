package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type availableModelsAdminService struct {
	*stubAdminService
	account service.Account
}

func (s *availableModelsAdminService) GetAccount(_ context.Context, id int64) (*service.Account, error) {
	if s.account.ID == id {
		acc := s.account
		return &acc, nil
	}
	return s.stubAdminService.GetAccount(context.Background(), id)
}

func setupAvailableModelsRouter(adminSvc service.AdminService, any2apiClient *service.Any2APIClient) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewAccountHandler(adminSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil, any2apiClient, nil, nil, nil)
	router.GET("/api/v1/admin/accounts/:id/models", handler.GetAvailableModels)
	router.POST("/api/v1/admin/accounts/:id/test", handler.Test)
	return router
}

func setupAny2APITestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/internal/providers/summary", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "ok",
			"service": map[string]any{"name": "any2api", "version": "test"},
			"models": map[string]any{
				"total": 4,
				"providers": map[string]int{
					"grok": 2,
					"qwen": 2,
				},
			},
			"accounts": map[string]any{
				"revision":   1,
				"total":      2,
				"manageable": 2,
				"selectable": 2,
				"pools": map[string]int{
					"default": 2,
				},
				"statuses": map[string]int{
					"active": 2,
				},
			},
		})
	})
	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"data": []map[string]any{
				{"id": "grok-4.20-0309", "object": "model", "owned_by": "grok", "display_name": "Grok Auto"},
				{"id": "grok-imagine-image-lite", "object": "model", "owned_by": "xai", "display_name": "Grok Image"},
				{"id": "qwen3.5-plus", "object": "model", "owned_by": "qwen", "display_name": "Qwen 3.5 Plus"},
				{"id": "qwen3.5-omni-plus", "object": "model", "owned_by": "alibaba", "display_name": "Qwen Omni"},
			},
		})
	})
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Model string `json:"model"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]any{
					"content": "ok:" + payload.Model,
				},
			}},
		})
	})
	return httptest.NewServer(mux)
}

func newAny2APIClientForTest(serverURL string) *service.Any2APIClient {
	return service.NewAny2APIClientWithSecret(service.Any2APISettings{
		Enabled:        true,
		BaseURL:        serverURL,
		TimeoutSeconds: 5,
	}, "test-key")
}

func TestAccountHandlerGetAvailableModels_OpenAIOAuthUsesExplicitModelMapping(t *testing.T) {
	svc := &availableModelsAdminService{
		stubAdminService: newStubAdminService(),
		account: service.Account{
			ID:       42,
			Name:     "openai-oauth",
			Platform: service.PlatformOpenAI,
			Type:     service.AccountTypeOAuth,
			Status:   service.StatusActive,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-5": "gpt-5.1",
				},
			},
		},
	}
	router := setupAvailableModelsRouter(svc, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts/42/models", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	require.Equal(t, "gpt-5", resp.Data[0].ID)
}

func TestAccountHandlerGetAvailableModels_OpenAIOAuthPassthroughFallsBackToDefaults(t *testing.T) {
	svc := &availableModelsAdminService{
		stubAdminService: newStubAdminService(),
		account: service.Account{
			ID:       43,
			Name:     "openai-oauth-passthrough",
			Platform: service.PlatformOpenAI,
			Type:     service.AccountTypeOAuth,
			Status:   service.StatusActive,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-5": "gpt-5.1",
				},
			},
			Extra: map[string]any{
				"openai_passthrough": true,
			},
		},
	}
	router := setupAvailableModelsRouter(svc, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts/43/models", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotEmpty(t, resp.Data)
	require.NotEqual(t, "gpt-5", resp.Data[0].ID)
}

func TestAccountHandlerGetAvailableModels_GrokAccountFallsBackToGrokDefaults(t *testing.T) {
	svc := &availableModelsAdminService{
		stubAdminService: newStubAdminService(),
		account: service.Account{
			ID:       44,
			Name:     "grok-apikey",
			Platform: service.PlatformGrok,
			Type:     service.AccountTypeAPIKey,
			Status:   service.StatusActive,
		},
	}
	router := setupAvailableModelsRouter(svc, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts/44/models", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotEmpty(t, resp.Data)
	require.Equal(t, "grok-4.20-0309-non-reasoning", resp.Data[0].ID)
}

func TestMaybeAppendAny2APIManagedAccounts_AppendsDistinctGrokAndQwenAccounts(t *testing.T) {
	server := setupAny2APITestServer(t)
	defer server.Close()

	handler := NewAccountHandler(newStubAdminService(), nil, nil, nil, nil, nil, nil, nil, nil, nil, newAny2APIClientForTest(server.URL), nil, nil, nil)
	accounts, total := handler.maybeAppendAny2APIManagedAccounts(context.Background(), nil, 0, 1, 20, "", "", "", "", 0)

	require.Len(t, accounts, 2)
	require.EqualValues(t, 2, total)
	require.NotEqual(t, accounts[0].ID, accounts[1].ID)
	platforms := []string{accounts[0].Platform, accounts[1].Platform}
	require.ElementsMatch(t, []string{service.PlatformGrok, service.PlatformQwen}, platforms)
}

func TestMaybeAppendAny2APIManagedAccounts_ReplacesProviderAlreadyPresent(t *testing.T) {
	server := setupAny2APITestServer(t)
	defer server.Close()

	handler := NewAccountHandler(newStubAdminService(), nil, nil, nil, nil, nil, nil, nil, nil, nil, newAny2APIClientForTest(server.URL), nil, nil, nil)
	existing := []service.Account{{
		ID:       1,
		Name:     "grok2api-upstream",
		Platform: service.PlatformGrok,
		Type:     service.AccountTypeAPIKey,
		Status:   service.StatusActive,
	}}
	accounts, total := handler.maybeAppendAny2APIManagedAccounts(context.Background(), existing, 1, 1, 20, "", "", "", "", 0)

	require.Len(t, accounts, 2)
	require.EqualValues(t, 2, total)
	require.Equal(t, syntheticAny2APIAccountIDGrok, accounts[0].ID)
	require.Equal(t, service.PlatformGrok, accounts[0].Platform)
	require.Equal(t, service.PlatformQwen, accounts[1].Platform)
}

func TestAccountHandlerGetAvailableModels_SyntheticAny2APIProviderFiltersModels(t *testing.T) {
	server := setupAny2APITestServer(t)
	defer server.Close()
	router := setupAvailableModelsRouter(newStubAdminService(), newAny2APIClientForTest(server.URL))

	tests := []struct {
		name      string
		accountID int64
		wantIDs   []string
	}{
		{name: "grok", accountID: syntheticAny2APIAccountIDGrok, wantIDs: []string{"grok-4.20-0309", "grok-imagine-image-lite"}},
		{name: "qwen", accountID: syntheticAny2APIAccountIDQwen, wantIDs: []string{"qwen3.5-plus", "qwen3.5-omni-plus"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts/"+strconv.FormatInt(tc.accountID, 10)+"/models", nil)
			router.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code)

			var resp struct {
				Data []struct {
					ID string `json:"id"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			ids := make([]string, 0, len(resp.Data))
			for _, item := range resp.Data {
				ids = append(ids, item.ID)
			}
			require.ElementsMatch(t, tc.wantIDs, ids)
		})
	}
}

func TestAccountHandlerGetAvailableModels_RealAny2APIManagedGrokAccountUsesAny2APIModels(t *testing.T) {
	server := setupAny2APITestServer(t)
	defer server.Close()
	svc := &availableModelsAdminService{
		stubAdminService: newStubAdminService(),
		account: service.Account{
			ID:       88,
			Name:     "grok2api-upstream",
			Platform: service.PlatformGrok,
			Type:     service.AccountTypeAPIKey,
			Status:   service.StatusActive,
		},
	}
	router := setupAvailableModelsRouter(svc, newAny2APIClientForTest(server.URL))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts/88/models", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	ids := make([]string, 0, len(resp.Data))
	for _, item := range resp.Data {
		ids = append(ids, item.ID)
	}
	require.ElementsMatch(t, []string{"grok-4.20-0309", "grok-imagine-image-lite"}, ids)
}

func TestAccountHandlerTest_SyntheticAny2APIProviderStreamsSuccessfulResult(t *testing.T) {
	server := setupAny2APITestServer(t)
	defer server.Close()
	router := setupAvailableModelsRouter(newStubAdminService(), newAny2APIClientForTest(server.URL))

	body := bytes.NewBufferString(`{"model_id":"grok-4.20-0309","prompt":"hello"}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/"+strconv.FormatInt(syntheticAny2APIAccountIDGrok, 10)+"/test", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"type":"test_start"`)
	require.Contains(t, rec.Body.String(), `"type":"test_complete"`)
	require.Contains(t, rec.Body.String(), `ok:grok-4.20-0309`)
}

func TestAccountHandlerTest_RealAny2APIManagedGrokAccountStreamsSuccessfulResult(t *testing.T) {
	server := setupAny2APITestServer(t)
	defer server.Close()
	svc := &availableModelsAdminService{
		stubAdminService: newStubAdminService(),
		account: service.Account{
			ID:       89,
			Name:     "grok2api-upstream",
			Platform: service.PlatformGrok,
			Type:     service.AccountTypeAPIKey,
			Status:   service.StatusActive,
		},
	}
	router := setupAvailableModelsRouter(svc, newAny2APIClientForTest(server.URL))

	body := bytes.NewBufferString(`{"model_id":"grok-4.20-0309","prompt":"hello"}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/89/test", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"type":"test_complete"`)
	require.Contains(t, rec.Body.String(), `ok:grok-4.20-0309`)
}
