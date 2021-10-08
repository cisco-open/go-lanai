package scheduler

import (
	"context"
	"time"
)

const (
	ModeFixedRate = iota
	ModeFixedDelay
	ModeRunOnce
)

type Mode int

type TaskFunc func(ctx context.Context) error

type TaskCanceller interface {
	Cancelled() <-chan error
	Cancel()
}

type TaskOptions func(opt *TaskOption) error
type TaskOption struct {
	name          string
	initialTime   time.Time
	interval      time.Duration
	mode          Mode
	cancelOnError bool
	hooks         []TaskHook
}

type TaskHook interface {
	// BeforeTrigger is invoked before each time the task is triggered. TaskHook can modify execution context
	BeforeTrigger(ctx context.Context, id string) context.Context

	// AfterTrigger is invoked after the task is triggered and executed
	AfterTrigger(ctx context.Context, id string, err error)
}
