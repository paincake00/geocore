package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// Event структура события, хранимая в памяти мок-сервера.
type Event struct {
	Body       json.RawMessage `json:"body"`
	ReceivedAt string          `json:"received_at"`
}

var (
	events []Event
	mu     sync.Mutex
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// POST: Принимаем вебхук
		if r.Method == http.MethodPost {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("Error reading body: %v", err)
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			log.Printf("Received Webhook: %s", string(body))

			// Check if body is valid JSON
			var js map[string]interface{}
			if json.Unmarshal(body, &js) != nil {
				log.Printf("Invalid JSON received")
			}

			mu.Lock()
			events = append(events, Event{
				Body:       json.RawMessage(body),
				ReceivedAt: time.Now().Format(time.RFC3339),
			})
			mu.Unlock()

			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "OK")
			return
		}

		// GET: Отдаем список полученных вебхуков
		if r.Method == http.MethodGet {
			mu.Lock()
			defer mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(events); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	log.Printf("Mock server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
