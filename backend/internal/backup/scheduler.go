package backup

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

const scheduledRunTimeout = 30 * time.Minute

// Scheduler runs the configured backup on a cron schedule, writing to BACKUP_BASE_DIR.
// It re-reads the singleton backup_config on every (re)load, so a config change just
// needs a Reload call.
type Scheduler struct {
	service  *Service
	cfgStore *store.BackupConfigStore
	mu       sync.Mutex
	cron     *cron.Cron
	baseCtx  context.Context
	cancel   context.CancelFunc
	started  bool
}

func NewScheduler(service *Service, cfgStore *store.BackupConfigStore) *Scheduler {
	return &Scheduler{service: service, cfgStore: cfgStore}
}

// Start initializes the scheduler's background context and loads the current config.
func (sc *Scheduler) Start(ctx context.Context) error {
	sc.mu.Lock()
	if sc.started {
		sc.mu.Unlock()
		return nil
	}
	sc.baseCtx, sc.cancel = context.WithCancel(context.Background())
	sc.started = true
	sc.mu.Unlock()

	return sc.Reload(ctx)
}

// Reload rebuilds the cron entry from the current backup_config. Safe to call after
// a config update. It is a no-op for scheduling when disabled or BACKUP_BASE_DIR is unset.
func (sc *Scheduler) Reload(ctx context.Context) error {
	cfg, err := sc.cfgStore.Get(ctx)
	if err != nil {
		return fmt.Errorf("load backup config: %w", err)
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()
	if !sc.started {
		return nil
	}

	if sc.cron != nil {
		sc.cron.Stop()
		sc.cron = nil
	}

	if !cfg.Enabled {
		return nil
	}
	if sc.service.baseDir == "" {
		log.Printf("backup scheduler: enabled but BACKUP_BASE_DIR is unset; not scheduling")
		return nil
	}

	c := cron.New()
	if _, err := c.AddFunc(cfg.Schedule, sc.runScheduled); err != nil {
		return fmt.Errorf("schedule backup %q: %w", cfg.Schedule, err)
	}
	c.Start()
	sc.cron = c
	return nil
}

// Stop cancels any in-flight run and waits for the scheduler to drain.
func (sc *Scheduler) Stop() {
	sc.mu.Lock()
	c := sc.cron
	cancel := sc.cancel
	sc.cron = nil
	sc.started = false
	sc.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if c != nil {
		<-c.Stop().Done()
	}
}

func (sc *Scheduler) runScheduled() {
	sc.mu.Lock()
	base := sc.baseCtx
	sc.mu.Unlock()
	if base == nil {
		return
	}

	ctx, cancel := context.WithTimeout(base, scheduledRunTimeout)
	defer cancel()

	cfg, err := sc.cfgStore.Get(ctx)
	if err != nil {
		log.Printf("backup scheduler: load config: %v", err)
		return
	}

	ranAt := time.Now()
	path, runErr := sc.service.Run(ctx, cfg.RetentionCount, ranAt)
	if runErr != nil {
		log.Printf("backup scheduler: run failed: %v", runErr)
	} else {
		log.Printf("backup scheduler: wrote %s", path)
	}

	if err := sc.cfgStore.UpdateRunStatus(ctx, ranAt, runErr); err != nil {
		log.Printf("backup scheduler: update run status: %v", err)
	}
}
