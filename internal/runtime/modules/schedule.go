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

// JobsCount returns the number of scheduled jobs
func (s *ScheduleModule) JobsCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.jobs)
}

// IsStarted returns whether the scheduler is running
func (s *ScheduleModule) IsStarted() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.started
}

// GetSchema implements JSSchemaProvider
func (s *ScheduleModule) GetSchema() JSModuleSchema {
	return JSModuleSchema{
		Name:        "schedule",
		Description: "Job scheduling for periodic tasks using cron expressions",
		Methods: []JSMethodSchema{
			{
				Name:        "daily",
				Description: "Schedule a job to run daily at midnight",
				Params:      []JSParamSchema{{Name: "handler", Type: "() => void", Description: "Function to execute"}},
			},
			{
				Name:        "hourly",
				Description: "Schedule a job to run at the start of every hour",
				Params:      []JSParamSchema{{Name: "handler", Type: "() => void", Description: "Function to execute"}},
			},
			{
				Name:        "cron",
				Description: "Schedule a job using a cron expression",
				Params: []JSParamSchema{
					{Name: "expression", Type: "string", Description: "Cron expression (e.g., '0 0 * * *')"},
					{Name: "handler", Type: "() => void", Description: "Function to execute"},
				},
			},
		},
	}
}

// GetScheduleSchema returns the schedule schema (static version)
func GetScheduleSchema() JSModuleSchema {
	return (&ScheduleModule{}).GetSchema()
}
