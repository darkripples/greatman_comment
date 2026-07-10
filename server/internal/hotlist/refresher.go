package hotlist

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"renwen/server/internal/runtime"
	"renwen/server/internal/zhihu"
)

const (
	checkEvery        = 1 * time.Minute
	refreshLimit      = 20
	secondLimitRetry  = 1 * time.Second
)

var (
	backoffMu     sync.Mutex
	dayLimitUntil time.Time
)

// StartRefresher periodically pulls Zhihu hot list into SQLite when min interval elapsed.
func StartRefresher(ctx context.Context, app *runtime.App) {
	go func() {
		refreshDue(ctx, app)
		ticker := time.NewTicker(checkEvery)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				refreshDue(ctx, app)
			}
		}
	}()
}

func refreshDue(ctx context.Context, app *runtime.App) {
	if shouldSkipDueToDayLimit() {
		return
	}

	cfg := app.Config()
	store := app.Store
	client := app.ZhihuClient()

	key := cacheKey(refreshLimit)
	entry, ok := store.GetAPICacheAny(key)
	if ok && time.Since(entry.FetchedAt) < cfg.HotListMinInterval {
		return
	}

	items, err := fetchHotListWithRetry(ctx, client, refreshLimit)
	if err != nil {
		if zhihu.IsDayLimitExceeded(err) {
			markDayLimit()
			log.Printf("[hot-list] refresh limit=%d skipped until %s: %v", refreshLimit, dayLimitUntil.Format(time.RFC3339), err)
			return
		}
		log.Printf("[hot-list] refresh limit=%d skipped: %v", refreshLimit, err)
		if !ok {
			_ = store.EnsureHotListSeedForKey(key)
		}
		return
	}
	if items == nil {
		items = []zhihu.HotItem{}
	}
	raw, err := json.Marshal(items)
	if err != nil {
		log.Printf("[hot-list] refresh limit=%d marshal: %v", refreshLimit, err)
		return
	}
	if err := store.SetAPICache(key, string(raw), cfg.HotListCacheTTL); err != nil {
		log.Printf("[hot-list] refresh limit=%d store: %v", refreshLimit, err)
		return
	}
	clearDayLimitBackoff()
	log.Printf("[hot-list] refreshed limit=%d (%d items)", refreshLimit, len(items))
}

func fetchHotListWithRetry(ctx context.Context, client *zhihu.Client, limit int) ([]zhihu.HotItem, error) {
	items, err := client.HotList(ctx, limit)
	if err == nil {
		return items, nil
	}
	if !zhihu.IsSecondLimitExceeded(err) {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(secondLimitRetry):
	}
	return client.HotList(ctx, limit)
}

func shouldSkipDueToDayLimit() bool {
	backoffMu.Lock()
	defer backoffMu.Unlock()
	if dayLimitUntil.IsZero() || time.Now().After(dayLimitUntil) {
		dayLimitUntil = time.Time{}
		return false
	}
	return true
}

func markDayLimit() {
	backoffMu.Lock()
	defer backoffMu.Unlock()
	now := time.Now()
	dayLimitUntil = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
}

func clearDayLimitBackoff() {
	backoffMu.Lock()
	defer backoffMu.Unlock()
	dayLimitUntil = time.Time{}
}

func cacheKey(limit int) string {
	return fmt.Sprintf("hot_list:limit=%d", limit)
}
