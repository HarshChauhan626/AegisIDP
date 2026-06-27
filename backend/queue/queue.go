package queue

import (
	"context"
	"errors"
)

// ErrQueueFull is returned when Enqueue is called on a full queue.
var ErrQueueFull = errors.New("job queue is at capacity")

// Job represents a unit of work to be processed by the workflow engine.
type Job struct {
	WorkflowID    string
	ExecutionType string         // e.g. "create-environment"
	EnvironmentID string
	Input         map[string]any // initial input data for the workflow
}

// Queue is a thread-safe, buffered in-memory job queue backed by a channel.
type Queue struct {
	ch chan Job
}

// New creates a Queue with the given buffer capacity.
func New(capacity int) *Queue {
	return &Queue{ch: make(chan Job, capacity)}
}

// Enqueue adds a job to the queue.
// Returns ErrQueueFull immediately if the queue is at capacity,
// or ctx.Err() if the context is cancelled.
func (q *Queue) Enqueue(ctx context.Context, job Job) error {
	select {
	case q.ch <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return ErrQueueFull
	}
}

// C returns the receive-only channel so workers can consume jobs.
func (q *Queue) C() <-chan Job {
	return q.ch
}

// Len returns the number of jobs currently waiting in the queue.
func (q *Queue) Len() int { return len(q.ch) }

// Cap returns the maximum queue capacity.
func (q *Queue) Cap() int { return cap(q.ch) }
