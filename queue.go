package goqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Queue represents a task queue
type Queue struct {
	redisClient *redis.Client
	queueName   string
}

// NewQueue initializes a task queue
func NewQueue(redisAddr string, queueName string) *Queue {
	client := redis.NewClient(&redis.Options{Addr: redisAddr})
	return &Queue{redisClient: client, queueName: queueName}
}

// Enqueue adds a task for immediate processing
func (q *Queue) Enqueue(taskName string, payload interface{}, priority Priority, maxRetries int) error {
	return q.scheduleTask(taskName, payload, priority, maxRetries, time.Now().Unix())
}

// ScheduleTask schedules a task for future execution
func (q *Queue) ScheduleTask(taskName string, payload interface{}, priority Priority, maxRetries int, runAt time.Time) error {
	return q.scheduleTask(taskName, payload, priority, maxRetries, runAt.Unix())
}

// scheduleTask pushes the task into Redis (immediate or delayed)
func (q *Queue) scheduleTask(taskName string, payload interface{}, priority Priority, maxRetries int, timestamp int64) error {
	ctx := context.Background()
	taskID := uuid.NewString()

	task := Task{
		ID:         taskID,
		Name:       taskName,
		Priority:   priority,
		MaxRetries: maxRetries,
		CreatedAt:  time.Now(),
	}

	taskData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task.Payload = taskData

	taskJSON, err := json.Marshal(task)
	if err != nil {
		return err
	}

	// Store the task in Redis Sorted Set with execution timestamp
	err = q.redisClient.ZAdd(ctx, fmt.Sprintf("%s:%s", q.queueName, priority), redis.Z{
		Score:  float64(timestamp),
		Member: taskJSON,
	}).Err()
	if err != nil {
		return err
	}

	return nil
}

// FetchScheduledTasks retrieves tasks that are due for execution

func (q *Queue) FetchScheduledTasks() (*Task, error) {
	ctx := context.Background()
	priorityQueues := []Priority{High, Medium, Low}

	for _, priority := range priorityQueues {
		queueName := fmt.Sprintf("%s:%s", q.queueName, priority) // Separate queues for each priority

		// Get tasks with timestamps <= current time
		now := float64(time.Now().Unix())
		taskJSONs, err := q.redisClient.ZRangeByScore(ctx, queueName, &redis.ZRangeBy{
			Min:   "-inf",
			Max:   fmt.Sprintf("%f", now),
			Count: 1,
		}).Result()

		if err == redis.Nil || len(taskJSONs) == 0 {
			continue // Try the next priority level
		} else if err != nil {
			return nil, err
		}

		// Attempt to remove the task atomically
		removed, err := q.redisClient.ZRem(ctx, queueName, taskJSONs[0]).Result()
		if err != nil {
			return nil, err
		}

		// Ensure the task was successfully removed
		if removed == 0 {
			continue // Task might have been taken by another worker
		}

		var task Task
		if err := json.Unmarshal([]byte(taskJSONs[0]), &task); err != nil {
			return nil, err
		}

		return &task, nil
	}

	return nil, nil // No task found in any priority queue
}
