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

// Worker отвечает за фоновую обработку задач (отправку вебхуков).
type Worker struct {
	Queue      usecase.QueueRepository
	QueueName  string
	WebhookURL string
	MaxRetries int
}

// New создает новый экземпляр воркера.
func New(q usecase.QueueRepository, webhookURL string) *Worker {
	return &Worker{
		Queue:      q,
		QueueName:  "webhook_tasks", // та же очередь, что и в сервисе
		WebhookURL: webhookURL,
		MaxRetries: 3,
	}
}

// Start запускает цикл обработки задач.
func (w *Worker) Start(ctx context.Context) {
	log.Println("Starting background worker...")
	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopped")
			return
		default:
			// Получаем задачу
			// Логика Dequeue в RedisRepo блокирует выполнение (timeout 0 = бесконечно).
			// Если контекст отменен, клиент Redis вернет ошибку.
			payloadJSON, err := w.Queue.Dequeue(ctx, w.QueueName)
			if err != nil {
				// Если ошибка из-за отмены контекста, выходим
				if ctx.Err() != nil {
					return
				}
				log.Printf("Worker dequeue error: %v", err)
				time.Sleep(1 * time.Second) // пауза при ошибке
				continue
			}

			// Обрабатываем асинхронно
			go w.processTask(payloadJSON)
		}
	}
}

// processTask обрабатывает одну задачу (отправку вебхука) с повторными попытками.
func (w *Worker) processTask(data string) {
	// Отправка на мок-сервер с ретраями
	log.Printf("Processing task: %s", data)

	// Здесь можно добавить валидацию JSON, но пока просто пересылаем.

	for i := 0; i < w.MaxRetries; i++ {
		err := w.sendWebhook(data)
		if err == nil {
			log.Printf("Webhook sent successfully")
			return
		}
		log.Printf("Failed to send webhook (attempt %d/%d): %v", i+1, w.MaxRetries, err)
		time.Sleep(time.Duration(2*i+1) * time.Second) // Линейная задержка: 1s, 3s, 5s...
	}
	log.Printf("Given up on task: %s", data)
}

// sendWebhook выполняет HTTP POST запрос на мок-сервер.
func (w *Worker) sendWebhook(data string) error {
	req, err := http.NewRequest("POST", w.WebhookURL, bytes.NewBufferString(data))
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
