package goqueue

import (
	"encoding/json"
	"time"
)

type Priority string

const (
	High   Priority = "high"
	Medium Priority = "medium"
	Low    Priority = "low"
)

// Task represents a job to be processed
type Task struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Payload    json.RawMessage `json:"payload"`
	Priority   Priority        `json:"priority"`
	MaxRetries int             `json:"max_retries"`
	CreatedAt  time.Time       `json:"created_at"`
}
