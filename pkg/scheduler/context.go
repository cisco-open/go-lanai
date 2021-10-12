package scheduler

import (
	"context"
	"time"
)

const (
	ModeFixedRate = iota
	ModeFixedDelay
	ModeRunOnce
	ModeDynamic
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
	mode          Mode
	initialTime   time.Time
	interval      time.Duration
	cancelOnError bool
	nextFunc      nextFunc
	hooks         []TaskHook
}

type TaskHook interface {
	// BeforeTrigger is invoked before each time the task is triggered. TaskHook can modify execution context
	BeforeTrigger(ctx context.Context, id string) context.Context

	// AfterTrigger is invoked after the task is triggered and executed
	AfterTrigger(ctx context.Context, id string, err error)
}

// nextFunc is used for ModeDynamic. It's currently unexported and only used for cron impl
type nextFunc func(time.Time) time.Time
