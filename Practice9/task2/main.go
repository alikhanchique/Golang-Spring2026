package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CachedResponse holds the stored result of a completed request.
type CachedResponse struct {
	StatusCode int
	Body       []byte
	Completed  bool
}

// MemoryStore is a thread-safe in-memory idempotency store.
type MemoryStore struct {
	mu   sync.Mutex
	data map[string]*CachedResponse
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string]*CachedResponse)}
}

func (m *MemoryStore) Get(key string) (*CachedResponse, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	resp, exists := m.data[key]
	return resp, exists
}

// StartProcessing atomically reserves the key. Returns false if already reserved.
func (m *MemoryStore) StartProcessing(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data[key]; exists {
		return false
	}
	m.data[key] = &CachedResponse{Completed: false}
	return true
}

func (m *MemoryStore) Finish(key string, status int, body []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if resp, exists := m.data[key]; exists {
		resp.StatusCode = status
		resp.Body = body
		resp.Completed = true
	} else {
		m.data[key] = &CachedResponse{StatusCode: status, Body: body, Completed: true}
	}
}

// IdempotencyMiddleware intercepts requests and deduplicates them by Idempotency-Key.
func IdempotencyMiddleware(store *MemoryStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			http.Error(w, "Idempotency-Key header required", http.StatusBadRequest)
			return
		}

		if cached, exists := store.Get(key); exists {
			if cached.Completed {
				fmt.Printf("[Middleware] Key %s already completed — returning cached result\n", key[:8])
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
			} else {
				fmt.Printf("[Middleware] Key %s is still processing — 409 Conflict\n", key[:8])
				http.Error(w, "Duplicate request in progress", http.StatusConflict)
			}
			return
		}

		if !store.StartProcessing(key) {
			if cached, exists := store.Get(key); exists && cached.Completed {
				fmt.Printf("[Middleware] Key %s just completed (race) — returning cached result\n", key[:8])
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
			} else {
				fmt.Printf("[Middleware] Key %s race — 409 Conflict\n", key[:8])
				http.Error(w, "Duplicate request in progress", http.StatusConflict)
			}
			return
		}

		fmt.Printf("[Middleware] Key %s is new — processing started\n", key[:8])
		recorder := httptest.NewRecorder()
		next.ServeHTTP(recorder, r)

		store.Finish(key, recorder.Code, recorder.Body.Bytes())

		for k, vals := range recorder.Header() {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(recorder.Code)
		w.Write(recorder.Body.Bytes())
	})
}

// paymentHandler simulates a heavy payment operation.
func paymentHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[Handler] Processing started — simulating 2s heavy work...")
	time.Sleep(2 * time.Second)

	txID := uuid.New().String()
	resp := map[string]interface{}{
		"status":         "paid",
		"amount":         1000,
		"transaction_id": txID,
	}
	body, _ := json.Marshal(resp)
	fmt.Printf("[Handler] Done — transaction_id: %s\n", txID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func main() {
	store := NewMemoryStore()
	mux := http.NewServeMux()
	mux.HandleFunc("/pay", paymentHandler)
	handler := IdempotencyMiddleware(store, mux)

	server := httptest.NewServer(handler)
	defer server.Close()

	idempotencyKey := uuid.New().String()
	fmt.Printf("\n=== Simulating double-click attack with key: %s ===\n\n", idempotencyKey[:8])

	const concurrentRequests = 7
	var wg sync.WaitGroup

	type result struct {
		goroutine int
		status    int
		body      string
	}
	results := make([]result, concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req, _ := http.NewRequest(http.MethodPost, server.URL+"/pay", nil)
			req.Header.Set("Idempotency-Key", idempotencyKey)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				results[idx] = result{goroutine: idx + 1, status: -1, body: err.Error()}
				return
			}
			defer resp.Body.Close()

			var buf []byte
			buf = make([]byte, 1024)
			n, _ := resp.Body.Read(buf)
			results[idx] = result{goroutine: idx + 1, status: resp.StatusCode, body: string(buf[:n])}
		}(i)
	}

	wg.Wait()

	fmt.Println("\n=== Results ===")
	for _, r := range results {
		fmt.Printf("Goroutine %d -> Status %d | Body: %s\n", r.goroutine, r.status, r.body)
	}

	fmt.Printf("\n=== Sending one more request after completion (same key: %s) ===\n", idempotencyKey[:8])
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/pay", nil)
	req.Header.Set("Idempotency-Key", idempotencyKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	var buf [1024]byte
	n, _ := resp.Body.Read(buf[:])
	fmt.Printf("Post-completion request -> Status %d | Body: %s\n", resp.StatusCode, string(buf[:n]))
}
