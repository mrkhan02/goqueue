package goqueue

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

type WorkerPool struct {
	queue         *Queue
	handlers      map[string]func(json.RawMessage) error
	workerCount   int
	wg            sync.WaitGroup
	resultBackend *ResultBackend
}

// NewWorkerPool initializes a worker pool with multiple workers
func NewWorkerPool(queue *Queue, workerCount int, resultBackend *ResultBackend) *WorkerPool {
	return &WorkerPool{
		queue:         queue,
		handlers:      make(map[string]func(json.RawMessage) error),
		workerCount:   workerCount,
		wg:            sync.WaitGroup{},
		resultBackend: resultBackend,
	}
}

// RegisterTask registers a task handler
func (wp *WorkerPool) RegisterTask(taskName string, handler func(json.RawMessage) error) {
	wp.handlers[taskName] = handler
}

// StartWorkers starts multiple workers in parallel
func (wp *WorkerPool) StartWorkers() {
	log.Printf("Starting %d workers...\n", wp.workerCount)

	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	wp.wg.Wait() // Wait for all workers to finish (this will run indefinitely)
}

// worker fetches and processes tasks
func (wp *WorkerPool) worker(workerID int) {
	defer wp.wg.Done()
	log.Printf("Worker %d started\n", workerID)

	for {
		task, err := wp.queue.FetchScheduledTasks()
		if err != nil {
			log.Printf("Worker %d: Error retrieving task: %v\n", workerID, err)
			continue
		}
		if task == nil {
			time.Sleep(100 * time.Millisecond) // No task available, wait a bit
			continue
		}

		log.Printf("Worker %d: Processing task %s\n", workerID, task.ID)
		wp.processTask(workerID, task, 0) // Process with retry
	}
}

// processTask handles retries
func (wp *WorkerPool) processTask(workerID int, task *Task, attempt int) {
	handler, exists := wp.handlers[task.Name]
	if !exists {
		log.Printf("Worker %d: No handler for task: %s\n", workerID, task.Name)
		return
	}

	err := handler(task.Payload)
	if err != nil {
		log.Printf("Worker %d: Task %s failed on attempt %d: %v\n", workerID, task.ID, attempt+1, err)

		if attempt < task.MaxRetries {
			backoff := time.Duration(2<<attempt) * time.Second
			time.Sleep(backoff)

			log.Printf("Worker %d: Retrying task %s after %v seconds\n", workerID, task.ID, backoff.Seconds())
			wp.processTask(workerID, task, attempt+1)
		} else {
			log.Printf("Worker %d: Task %s failed after max retries\n", workerID, task.ID)
			wp.resultBackend.StoreResult(task.ID, fmt.Sprintf("%s:%s:%s", "failed", "max retries exceeded", err.Error()))

		}
	} else {
		log.Printf("Worker %d: Task %s completed successfully\n", workerID, task.ID)
		wp.resultBackend.StoreResult(task.ID, "success")
	}
}
