package scheduler

import (
	"context"
	"testing"
	"time"
)

// Mock job for testing
type MockJob struct {
	name     string
	runCount int
}

func (j *MockJob) Name() string {
	return j.name
}

func (j *MockJob) Run(ctx context.Context) error {
	j.runCount++
	return nil
}

func TestScheduler(t *testing.T) {
	s := NewScheduler()
	mockJob := &MockJob{name: "test_job"}

	// Test adding a job
	err := s.AddJob("* * * * * *", mockJob) // Run every second
	if err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	// Start the scheduler
	s.Start()
	defer s.Stop()

	// Wait for the job to run at least once
	time.Sleep(2 * time.Second)

	// Verify the job ran
	if mockJob.runCount == 0 {
		t.Error("Job did not run")
	}

	// Test running a job now
	initialRunCount := mockJob.runCount
	err = s.RunJobNow("test_job")
	if err != nil {
		t.Fatalf("Failed to run job now: %v", err)
	}

	if mockJob.runCount != initialRunCount+1 {
		t.Errorf("RunJobNow did not increment run count")
	}

	// Test running a non-existent job
	err = s.RunJobNow("non_existent_job")
	if err == nil {
		t.Error("Running non-existent job should have failed")
	}
}

func TestAddMorningEveningJob(t *testing.T) {
	s := NewScheduler()
	mockJob := &MockJob{name: "test_morning_evening_job"}

	// Test adding a morning/evening job
	err := s.AddJob("0 0 10 * * *", mockJob) // Just add one schedule to test
	if err != nil {
		t.Fatalf("Failed to add morning job: %v", err)
	}

	// Verify the job is registered
	err = s.RunJobNow("test_morning_evening_job")
	if err != nil {
		t.Errorf("Morning job not registered correctly: %v", err)
	}
}
