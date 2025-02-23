package goqueue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// ResultBackend represents the result backend
type ResultBackend struct {
	redisClient *redis.Client
}

// NewResultBackend initializes the result backend
func NewResultBackend(redisAddr string) *ResultBackend {
	client := redis.NewClient(&redis.Options{Addr: redisAddr})
	return &ResultBackend{redisClient: client}
}

// StoreResult saves the task result
func (rb *ResultBackend) StoreResult(taskID string, result interface{}) error {
	ctx := context.Background()
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return rb.redisClient.Set(ctx, fmt.Sprintf("task:%s", taskID), resultJSON, 0).Err()
}

// GetResult retrieves the task result
func (rb *ResultBackend) GetResult(taskID string) (string, error) {
	ctx := context.Background()
	result, err := rb.redisClient.Get(ctx, fmt.Sprintf("task:%s", taskID)).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return result, nil
}
