package scheduler

import (
	"context"
	"sync"
	"time"

	"core/internal/database"
	"core/internal/imap"
	"github.com/rs/zerolog/log"
)

type Scheduler struct {
	db           *database.DB
	syncer       *imap.Syncer
	maxWorkers   int
	batchSize    int
	syncInterval time.Duration
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

func NewScheduler(db *database.DB, syncer *imap.Syncer, maxWorkers, batchSize int, syncInterval time.Duration) *Scheduler {
	return &Scheduler{
		db: db, syncer: syncer, maxWorkers: maxWorkers,
		batchSize: batchSize, syncInterval: syncInterval,
		stopChan: make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	log.Info().Msgf("Starting scheduler: %d workers, batch %d, interval %v", s.maxWorkers, s.batchSize, s.syncInterval)
	s.wg.Add(1)
	go s.run()
}

func (s *Scheduler) Stop() {
	log.Info().Msg("Stopping scheduler")
	close(s.stopChan)
	s.wg.Wait()
	log.Info().Msg("Scheduler stopped")
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
	log.Info().Msg("Sync cycle started")

	integrations, err := s.db.GetIntegrationsForSync(ctx, s.batchSize)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get integrations")
		return
	}

	if len(integrations) == 0 {
		return
	}

	log.Info().Msgf("Found %d integrations", len(integrations))

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
			log.Error().Err(err).Msg("Sync failed")
		} else {
			success++
		}
	}

	log.Info().Msgf("Sync completed: %d/%d", success, len(integrations))
}

func (s *Scheduler) worker(jobs <-chan database.EmailIntegration, results chan<- error) {
	for integration := range jobs {
		results <- s.syncer.SyncIntegration(&integration)
	}
}
