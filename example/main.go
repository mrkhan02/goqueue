package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	taskqueue "github.com/mrkhan02/goqueue"
)

func main() {
	// Initialize queue
	queue := taskqueue.NewQueue("localhost:6379", "tasks")

	// Initialize result backend
	resultBackend := taskqueue.NewResultBackend("localhost:6379")
	// Example task payload
	payload := map[string]string{"email": "user@example.com"}

	// Enqueue a task
	err := queue.Enqueue("send_email", payload, taskqueue.High, 3)
	if err != nil {
		log.Fatal(err)
	}
	// Initialize worker
	worker := taskqueue.NewWorkerPool(queue, 1, resultBackend)

	// Register task handlers
	worker.RegisterTask("send_email", func(payload json.RawMessage) error {
		var data map[string]string
		if err := json.Unmarshal(payload, &data); err != nil {
			return err
		}
		fmt.Println("📧 Sending email to:", data["email"])
		time.Sleep(2 * time.Second) // Simulate processing
		return nil
	})

	// Start worker
	worker.StartWorkers()
}
