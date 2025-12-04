package modules

import (
	"log/slog"
	"sync"

	"github.com/dop251/goja"
	"github.com/robfig/cron/v3"
)

type ScheduleModule struct {
	cron    *cron.Cron
	jobs    []cron.EntryID
	mu      sync.Mutex
	started bool
	logger  *slog.Logger
}

func NewScheduleModule(logger *slog.Logger) *ScheduleModule {
	return &ScheduleModule{
		cron:   cron.New(),
		jobs:   []cron.EntryID{},
		logger: logger,
	}
}

func (s *ScheduleModule) addJob(spec string, handler goja.Callable) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, err := s.cron.AddFunc(spec, func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("Scheduled job panic", "error", r)
			}
		}()

		_, err := handler(nil, nil)
		if err != nil {
			s.logger.Error("Scheduled job error", "error", err)
		}
	})

	if err != nil {
		s.logger.Error("Failed to add scheduled job", "spec", spec, "error", err)
		return
	}

	s.jobs = append(s.jobs, id)
}

func (s *ScheduleModule) Daily(handler goja.Callable) {
	// Run at midnight every day
	s.addJob("0 0 * * *", handler)
}

func (s *ScheduleModule) Hourly(handler goja.Callable) {
	// Run at the start of every hour
	s.addJob("0 * * * *", handler)
}

func (s *ScheduleModule) Cron(expression string, handler goja.Callable) {
	s.addJob(expression, handler)
}

func (s *ScheduleModule) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started && len(s.jobs) > 0 {
		s.cron.Start()
		s.started = true
		s.logger.Info("Scheduler started", "jobs", len(s.jobs))
	}
}

func (s *ScheduleModule) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		s.cron.Stop()
		s.started = false
		s.logger.Info("Scheduler stopped")
	}
}
