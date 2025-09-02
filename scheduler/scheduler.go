package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

// Job represents a scheduled job
type Job interface {
	Name() string
	Run(ctx context.Context) error
}

// Scheduler manages scheduled jobs
type Scheduler struct {
	cron      *cron.Cron
	jobs      map[string]Job
	isRunning bool
}

// NewScheduler creates a new scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{
		cron: cron.New(
			cron.WithSeconds(),
			cron.WithLogger(cron.VerbosePrintfLogger(log.Default())),
			cron.WithChain(cron.Recover(cron.DefaultLogger)),
		),
		jobs: make(map[string]Job),
	}
}

// AddJob adds a job to the scheduler with a cron specification
func (s *Scheduler) AddJob(spec string, job Job) error {
	name := job.Name()
	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job %s already registered", name)
	}

	_, err := s.cron.AddFunc(spec, func() {
		log.Printf("Starting scheduled job: %s", name)
		startTime := time.Now()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		if err := job.Run(ctx); err != nil {
			log.Printf("Error running job %s: %v", name, err)
		} else {
			duration := time.Since(startTime)
			log.Printf("Completed job %s in %s", name, duration)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to add job %s: %v", name, err)
	}

	s.jobs[name] = job
	return nil
}

// AddMorningEveningJob adds a job that runs at 10am and 5pm every day
func (s *Scheduler) AddMorningEveningJob(job Job) error {
	// 10am every day: 0 0 10 * * *
	// 5pm every day: 0 0 17 * * *
	if err := s.AddJob("0 0 10 * * *", job); err != nil {
		return err
	}
	return s.AddJob("0 0 17 * * *", job)
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	if s.isRunning {
		return
	}
	s.cron.Start()
	s.isRunning = true
	log.Println("Scheduler started")
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	if !s.isRunning {
		return
	}
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.isRunning = false
	log.Println("Scheduler stopped")
}

// RunJobNow runs a job immediately outside of schedule
func (s *Scheduler) RunJobNow(name string) error {
	job, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job %s not registered", name)
	}

	log.Printf("Manually running job: %s", name)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	return job.Run(ctx)
}
