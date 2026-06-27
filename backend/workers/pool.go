package workers

import (
	"context"
	"sync"

	"github.com/HarshChauhan626/AegisIDP/backend/queue"
	"github.com/HarshChauhan626/AegisIDP/backend/workflow"
	"go.uber.org/zap"
)

// WorkerPool manages a fixed number of goroutines that consume jobs from a Queue
// and execute them via the workflow Engine.
type WorkerPool struct {
	workerCount int
	queue       *queue.Queue
	engine      *workflow.Engine
	log         *zap.Logger

	// cancels maps workflowID → context.CancelFunc so the Cancel endpoint
	// can signal a specific running workflow.
	cancels sync.Map
	wg      sync.WaitGroup
}

// NewWorkerPool creates a WorkerPool. Call Start to begin processing.
func NewWorkerPool(n int, q *queue.Queue, engine *workflow.Engine, log *zap.Logger) *WorkerPool {
	return &WorkerPool{
		workerCount: n,
		queue:       q,
		engine:      engine,
		log:         log,
	}
}

// Start launches all worker goroutines. They run until ctx is cancelled.
func (p *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.run(ctx, i)
	}
	p.log.Info("worker pool started", zap.Int("workers", p.workerCount))
}

// Stop blocks until all in-flight jobs finish after the parent context is cancelled.
func (p *WorkerPool) Stop() {
	p.wg.Wait()
	p.log.Info("worker pool stopped")
}

// Cancel signals cancellation for a specific workflow run.
// Returns true if a running context was found and cancelled, false otherwise.
func (p *WorkerPool) Cancel(workflowID string) bool {
	if v, ok := p.cancels.Load(workflowID); ok {
		v.(context.CancelFunc)()
		return true
	}
	return false
}

// run is the main loop for a single worker goroutine.
func (p *WorkerPool) run(ctx context.Context, id int) {
	defer p.wg.Done()
	log := p.log.With(zap.Int("worker_id", id))
	log.Info("worker started")

	for {
		select {
		case <-ctx.Done():
			log.Info("worker shutting down")
			return
		case job := <-p.queue.C():
			p.processJob(ctx, job, log)
		}
	}
}

// processJob creates a cancellable child context, registers it, then runs the engine.
func (p *WorkerPool) processJob(parentCtx context.Context, job queue.Job, log *zap.Logger) {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	// Register so Cancel(workflowID) can signal this context
	p.cancels.Store(job.WorkflowID, cancel)
	defer p.cancels.Delete(job.WorkflowID)

	log.Info("processing job",
		zap.String("workflow_id", job.WorkflowID),
		zap.String("type", job.ExecutionType),
		zap.String("environment_id", job.EnvironmentID),
	)

	if err := p.engine.Run(ctx, job.WorkflowID, job.ExecutionType, job.EnvironmentID, job.Input); err != nil {
		log.Error("job finished with error",
			zap.String("workflow_id", job.WorkflowID),
			zap.Error(err),
		)
	} else {
		log.Info("job finished successfully", zap.String("workflow_id", job.WorkflowID))
	}
}
