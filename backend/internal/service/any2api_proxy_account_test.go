package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type any2APIProxyAccountRepoStub struct {
	findResults []Account
	findErr     error
	createErr   error
	created     []*Account
	nextID      int64
}

func (s *any2APIProxyAccountRepoStub) Create(ctx context.Context, account *Account) error {
	if s.createErr != nil {
		return s.createErr
	}
	s.nextID++
	account.ID = s.nextID
	copyAccount := *account
	s.created = append(s.created, &copyAccount)
	return nil
}

func (s *any2APIProxyAccountRepoStub) GetByID(ctx context.Context, id int64) (*Account, error) {
	panic("unexpected GetByID call")
}

func (s *any2APIProxyAccountRepoStub) GetByIDs(ctx context.Context, ids []int64) ([]*Account, error) {
	panic("unexpected GetByIDs call")
}

func (s *any2APIProxyAccountRepoStub) ExistsByID(ctx context.Context, id int64) (bool, error) {
	panic("unexpected ExistsByID call")
}

func (s *any2APIProxyAccountRepoStub) GetByCRSAccountID(ctx context.Context, crsAccountID string) (*Account, error) {
	panic("unexpected GetByCRSAccountID call")
}

func (s *any2APIProxyAccountRepoStub) FindByExtraField(ctx context.Context, key string, value any) ([]Account, error) {
	if s.findErr != nil {
		return nil, s.findErr
	}
	results := make([]Account, len(s.findResults))
	copy(results, s.findResults)
	return results, nil
}

func (s *any2APIProxyAccountRepoStub) ListCRSAccountIDs(ctx context.Context) (map[string]int64, error) {
	panic("unexpected ListCRSAccountIDs call")
}

func (s *any2APIProxyAccountRepoStub) Update(ctx context.Context, account *Account) error {
	panic("unexpected Update call")
}

func (s *any2APIProxyAccountRepoStub) Delete(ctx context.Context, id int64) error {
	panic("unexpected Delete call")
}

func (s *any2APIProxyAccountRepoStub) List(ctx context.Context, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (s *any2APIProxyAccountRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, platform, accountType, status, search string, groupID int64, privacyMode string) ([]Account, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (s *any2APIProxyAccountRepoStub) ListByGroup(ctx context.Context, groupID int64) ([]Account, error) {
	panic("unexpected ListByGroup call")
}

func (s *any2APIProxyAccountRepoStub) ListActive(ctx context.Context) ([]Account, error) {
	panic("unexpected ListActive call")
}

func (s *any2APIProxyAccountRepoStub) ListByPlatform(ctx context.Context, platform string) ([]Account, error) {
	panic("unexpected ListByPlatform call")
}

func (s *any2APIProxyAccountRepoStub) UpdateLastUsed(ctx context.Context, id int64) error {
	panic("unexpected UpdateLastUsed call")
}

func (s *any2APIProxyAccountRepoStub) BatchUpdateLastUsed(ctx context.Context, updates map[int64]time.Time) error {
	panic("unexpected BatchUpdateLastUsed call")
}

func (s *any2APIProxyAccountRepoStub) SetError(ctx context.Context, id int64, errorMsg string) error {
	panic("unexpected SetError call")
}

func (s *any2APIProxyAccountRepoStub) ClearError(ctx context.Context, id int64) error {
	panic("unexpected ClearError call")
}

func (s *any2APIProxyAccountRepoStub) SetSchedulable(ctx context.Context, id int64, schedulable bool) error {
	panic("unexpected SetSchedulable call")
}

func (s *any2APIProxyAccountRepoStub) AutoPauseExpiredAccounts(ctx context.Context, now time.Time) (int64, error) {
	panic("unexpected AutoPauseExpiredAccounts call")
}

func (s *any2APIProxyAccountRepoStub) BindGroups(ctx context.Context, accountID int64, groupIDs []int64) error {
	panic("unexpected BindGroups call")
}

func (s *any2APIProxyAccountRepoStub) ListSchedulable(ctx context.Context) ([]Account, error) {
	panic("unexpected ListSchedulable call")
}

func (s *any2APIProxyAccountRepoStub) ListSchedulableByGroupID(ctx context.Context, groupID int64) ([]Account, error) {
	panic("unexpected ListSchedulableByGroupID call")
}

func (s *any2APIProxyAccountRepoStub) ListSchedulableByPlatform(ctx context.Context, platform string) ([]Account, error) {
	panic("unexpected ListSchedulableByPlatform call")
}

func (s *any2APIProxyAccountRepoStub) ListSchedulableByGroupIDAndPlatform(ctx context.Context, groupID int64, platform string) ([]Account, error) {
	panic("unexpected ListSchedulableByGroupIDAndPlatform call")
}

func (s *any2APIProxyAccountRepoStub) ListSchedulableByPlatforms(ctx context.Context, platforms []string) ([]Account, error) {
	panic("unexpected ListSchedulableByPlatforms call")
}

func (s *any2APIProxyAccountRepoStub) ListSchedulableByGroupIDAndPlatforms(ctx context.Context, groupID int64, platforms []string) ([]Account, error) {
	panic("unexpected ListSchedulableByGroupIDAndPlatforms call")
}

func (s *any2APIProxyAccountRepoStub) ListSchedulableUngroupedByPlatform(ctx context.Context, platform string) ([]Account, error) {
	panic("unexpected ListSchedulableUngroupedByPlatform call")
}

func (s *any2APIProxyAccountRepoStub) ListSchedulableUngroupedByPlatforms(ctx context.Context, platforms []string) ([]Account, error) {
	panic("unexpected ListSchedulableUngroupedByPlatforms call")
}

func (s *any2APIProxyAccountRepoStub) SetRateLimited(ctx context.Context, id int64, resetAt time.Time) error {
	panic("unexpected SetRateLimited call")
}

func (s *any2APIProxyAccountRepoStub) SetModelRateLimit(ctx context.Context, id int64, scope string, resetAt time.Time) error {
	panic("unexpected SetModelRateLimit call")
}

func (s *any2APIProxyAccountRepoStub) SetOverloaded(ctx context.Context, id int64, until time.Time) error {
	panic("unexpected SetOverloaded call")
}

func (s *any2APIProxyAccountRepoStub) SetTempUnschedulable(ctx context.Context, id int64, until time.Time, reason string) error {
	panic("unexpected SetTempUnschedulable call")
}

func (s *any2APIProxyAccountRepoStub) ClearTempUnschedulable(ctx context.Context, id int64) error {
	panic("unexpected ClearTempUnschedulable call")
}

func (s *any2APIProxyAccountRepoStub) ClearRateLimit(ctx context.Context, id int64) error {
	panic("unexpected ClearRateLimit call")
}

func (s *any2APIProxyAccountRepoStub) ClearAntigravityQuotaScopes(ctx context.Context, id int64) error {
	panic("unexpected ClearAntigravityQuotaScopes call")
}

func (s *any2APIProxyAccountRepoStub) ClearModelRateLimits(ctx context.Context, id int64) error {
	panic("unexpected ClearModelRateLimits call")
}

func (s *any2APIProxyAccountRepoStub) UpdateSessionWindow(ctx context.Context, id int64, start, end *time.Time, status string) error {
	panic("unexpected UpdateSessionWindow call")
}

func (s *any2APIProxyAccountRepoStub) UpdateExtra(ctx context.Context, id int64, updates map[string]any) error {
	panic("unexpected UpdateExtra call")
}

func (s *any2APIProxyAccountRepoStub) BulkUpdate(ctx context.Context, ids []int64, updates AccountBulkUpdate) (int64, error) {
	panic("unexpected BulkUpdate call")
}

func (s *any2APIProxyAccountRepoStub) IncrementQuotaUsed(ctx context.Context, id int64, amount float64) error {
	panic("unexpected IncrementQuotaUsed call")
}

func (s *any2APIProxyAccountRepoStub) ResetQuotaUsed(ctx context.Context, id int64) error {
	panic("unexpected ResetQuotaUsed call")
}

func TestOpenAIGatewayServiceEnsureAny2APIProxyAccount_ReusesExisting(t *testing.T) {
	repo := &any2APIProxyAccountRepoStub{
		findResults: []Account{{
			ID:       44,
			Name:     "grok2api-upstream",
			Platform: PlatformGrok,
			Extra: map[string]any{
				any2APIProxyAccountExtraKey: "any2api:grok",
			},
		}},
	}
	svc := &OpenAIGatewayService{accountRepo: repo}

	account, err := svc.EnsureAny2APIProxyAccount(context.Background(), "grok-4.20-0309")
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, int64(44), account.ID)
	require.Empty(t, repo.created)
}

func TestOpenAIGatewayServiceEnsureAny2APIProxyAccount_CreatesSyntheticUsageAccount(t *testing.T) {
	repo := &any2APIProxyAccountRepoStub{nextID: 1000}
	svc := &OpenAIGatewayService{accountRepo: repo}

	account, err := svc.EnsureAny2APIProxyAccount(context.Background(), "qwen3.5-plus")
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Len(t, repo.created, 1)
	require.Equal(t, int64(1001), account.ID)
	require.Equal(t, "qwen2api-upstream", account.Name)
	require.Equal(t, PlatformQwen, account.Platform)
	require.Equal(t, AccountTypeAPIKey, account.Type)
	require.False(t, account.Schedulable)
	require.Equal(t, StatusActive, account.Status)
	require.Equal(t, "any2api:qwen", account.Extra[any2APIProxyAccountExtraKey])
	require.NotNil(t, account.Credentials)
	require.Empty(t, account.Credentials)
}

var _ AccountRepository = (*any2APIProxyAccountRepoStub)(nil)
