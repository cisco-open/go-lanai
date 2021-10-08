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
}

