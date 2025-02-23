# goqueue

### Overview
This package provides a distributed task queue system similar to [Celery](https://docs.celeryq.dev/en/stable/) but implemented in **Golang**. It enables executing background jobs asynchronously, supports task retries with exponential backoff, priority queues.

### Features
- **Distributed task execution**
- **Retry mechanism with exponential backoff**
- **Task scheduling**
- **Priority-based task queuing**
- **Redis-based task persistence**

### Installation
```sh
go get -u github.com/mrkhan02/goqueue
```

### Usage
#### 1. Import Package
```go
import (
    "github.com/mrkhan02/goqueue"
    "encoding/json"
    "log"
)
```

#### 2. Initialize Task Queue
```go
queue := taskqueue.NewQueue("localhost:6379", "tasks")
resultBackend := taskqueue.NewResultBackend("localhost:6379") // store result
workerPool := taskqueue.NewWorkerPool(queue, 5,resultBackend) // 5 concurrent workers
```

#### 3. Register Task Handlers
```go
workerPool.RegisterTask("send_email", func(payload json.RawMessage) error {
    var data map[string]string
    if err := json.Unmarshal(payload, &data); err != nil {
        return err
    }
    log.Printf("Sending email to: %s", data["email"])
    return nil
})
```

#### 4. Enqueue Tasks
```go
payload := map[string]string{"email": "user@example.com"}
queue.Enqueue("send_email", payload, taskqueue.High, 3)
```
- `High` priority task
- Max retries: **3**

#### 5. Start Worker Pool
```go
workerPool.StartWorkers()
```

### Architecture
1. **Producer**: Pushes tasks to Redis queue.
2. **Redis Queue**: Stores task data persistently.
3. **Worker Pool**: Fetches tasks and executes them.
4. **Retry Queue**: Handles failed tasks using exponential backoff.

### Task Retries
- If a task fails, it is retried **up to max retries**.
- Uses **exponential backoff** (`2^retry_count` seconds).

### License
MIT License.

---

🚀 **Start processing tasks asynchronously in Go!**

