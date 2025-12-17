package scheduler

import (
	"context"
	"sync"
	"time"

	"core/internal/database"
	"core/internal/imap"
	"reminder-hub/pkg/logger"
)

type Scheduler struct {
	db           *database.DB
	syncer       *imap.Syncer
	maxWorkers   int
	batchSize    int
	syncInterval time.Duration
	stopChan     chan struct{}
	wg           sync.WaitGroup
	log          *logger.CurrentLogger
}

func NewScheduler(db *database.DB, syncer *imap.Syncer, maxWorkers, batchSize int, syncInterval time.Duration, log *logger.CurrentLogger) *Scheduler {
	return &Scheduler{
		db: db, syncer: syncer, maxWorkers: maxWorkers,
		batchSize: batchSize, syncInterval: syncInterval,
		stopChan: make(chan struct{}),
		log:      log,
	}
}

func (s *Scheduler) Start() {
	ctx := context.Background()
	s.log.Info(ctx, "Starting scheduler", "workers", s.maxWorkers, "batch", s.batchSize, "interval", s.syncInterval.String())
	s.wg.Add(1)
	go s.run()
}

func (s *Scheduler) Stop() {
	ctx := context.Background()
	s.log.Info(ctx, "Stopping scheduler")
	close(s.stopChan)
	s.wg.Wait()
	s.log.Info(ctx, "Scheduler stopped")
}

func (s *Scheduler) run() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	s.syncAll()
	for {
		select {
		case <-ticker.C:
			s.syncAll()
		case <-s.stopChan:
			return
		}
	}
}

func (s *Scheduler) syncAll() {
	ctx := context.Background()
	s.log.Info(ctx, "Sync cycle started")

	integrations, err := s.db.GetIntegrationsForSync(ctx, s.batchSize)
	if err != nil {
		s.log.Error(ctx, "Failed to get integrations", "error", err)
		return
	}

	if len(integrations) == 0 {
		return
	}

	s.log.Info(ctx, "Found integrations", "count", len(integrations))

	jobs := make(chan database.EmailIntegration, len(integrations))
	results := make(chan error, len(integrations))

	for i := 0; i < s.maxWorkers; i++ {
		go s.worker(jobs, results)
	}

	for _, integration := range integrations {
		jobs <- integration
	}
	close(jobs)

	success := 0
	for range integrations {
		if err := <-results; err != nil {
			s.log.Error(ctx, "Sync failed", "error", err)
		} else {
			success++
		}
	}

	s.log.Info(ctx, "Sync completed", "success", success, "total", len(integrations))
}

func (s *Scheduler) worker(jobs <-chan database.EmailIntegration, results chan<- error) {
	for integration := range jobs {
		results <- s.syncer.SyncIntegration(&integration)
	}
}
