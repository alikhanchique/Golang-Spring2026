package exchange

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *ExchangeService) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(func() { srv.Close() })
	svc := NewExchangeService(srv.URL)
	return srv, svc
}

func TestGetRate_Success(t *testing.T) {
	_, svc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/convert", r.URL.Path)
		assert.Equal(t, "USD", r.URL.Query().Get("from"))
		assert.Equal(t, "EUR", r.URL.Query().Get("to"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"base":"USD","target":"EUR","rate":0.92}`))
	})

	rate, err := svc.GetRate("USD", "EUR")
	assert.NoError(t, err)
	assert.InDelta(t, 0.92, rate, 0.0001)
}

func TestGetRate_APIBusinessError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"404 not found", http.StatusNotFound},
		{"400 bad request", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, svc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(`{"error":"invalid currency pair"}`))
			})

			_, err := svc.GetRate("USD", "XYZ")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid currency pair")
		})
	}
}

func TestGetRate_MalformedJSON(t *testing.T) {
	_, svc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Send invalid JSON
		_, _ = w.Write([]byte(`Internal Server Error`))
	})

	_, err := svc.GetRate("USD", "EUR")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode error")
}

func TestGetRate_Timeout(t *testing.T) {
	_, svc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})

	svc.Client.Timeout = 100 * time.Millisecond

	_, err := svc.GetRate("USD", "EUR")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network error")
}

func TestGetRate_ServerPanic(t *testing.T) {
	_, svc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal server error"}`))
	})

	_, err := svc.GetRate("USD", "EUR")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api error")
}
func TestGetRate_EmptyBody(t *testing.T) {
	_, svc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Write nothing — empty body
	})

	_, err := svc.GetRate("USD", "EUR")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode error")
}

func TestGetRate_TruncatedJSON(t *testing.T) {
	_, svc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Truncated — not a valid JSON object
		_, _ = w.Write([]byte(`{"base":"USD","rate":`))
	})

	_, err := svc.GetRate("USD", "EUR")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode error")
}

func TestGetRate_UnexpectedStatus_NoErrorBody(t *testing.T) {
	_, svc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTeapot) // 418
		_, _ = w.Write([]byte(`{"base":"USD","target":"EUR","rate":0}`))
	})

	_, err := svc.GetRate("USD", "EUR")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status")

	assert.True(t, strings.Contains(err.Error(), "418"))
}
