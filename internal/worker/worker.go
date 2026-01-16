package worker

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/paincake00/geocore/internal/usecase"
)

type Worker struct {
	Queue         usecase.QueueRepository
	QueueName     string
	MockServerURL string
	MaxRetries    int
}

func New(q usecase.QueueRepository, mockURL string) *Worker {
	return &Worker{
		Queue:         q,
		QueueName:     "webhook_tasks", // same as in Service
		MockServerURL: mockURL,
		MaxRetries:    3,
	}
}

func (w *Worker) Start(ctx context.Context) {
	log.Println("Starting background worker...")
	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopped")
			return
		default:
			// Fetch task
			// Dequeue logic in RedisRepo should block for a bit (0 is infinite block).
			// If we want graceful shutdown we need to handle that in Dequeue implementation or here.
			// Current Dequeue implementation uses BRPop with 0 timeout (infinite).
			// If context is canceled, client.BRPop should return error.
			payloadJSON, err := w.Queue.Dequeue(ctx, w.QueueName)
			if err != nil {
				// If error is due to context cancel, we exit
				if ctx.Err() != nil {
					return
				}
				log.Printf("Worker dequeue error: %v", err)
				time.Sleep(1 * time.Second) // backoff on error
				continue
			}

			// Process async
			go w.processTask(payloadJSON)
		}
	}
}

func (w *Worker) processTask(data string) {
	// Send to Mock Server with retry
	log.Printf("Processing task: %s", data)

	// Since we already have the JSON string, we can just send it.
	// But let's validate or parse if needed. For now just raw forward.

	for i := 0; i < w.MaxRetries; i++ {
		err := w.sendWebhook(data)
		if err == nil {
			log.Printf("Webhook sent successfully")
			return
		}
		log.Printf("Failed to send webhook (attempt %d/%d): %v", i+1, w.MaxRetries, err)
		time.Sleep(time.Duration(2*i+1) * time.Second) // Linear backoff: 1s, 3s, 5s...
	}
	log.Printf("Given up on task: %s", data)
}

func (w *Worker) sendWebhook(data string) error {
	req, err := http.NewRequest("POST", w.MockServerURL, bytes.NewBufferString(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}
	return nil
}
