package practice8

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type RateResponse struct {
	Base     string  `json:"base"`
	Target   string  `json:"target"`
	Rate     float64 `json:"rate"`
	ErrorMsg string  `json:"error,omitempty"`
}

type ExchangeService struct {
	BaseURL string
	Client  *http.Client
}

func NewExchangeService(baseURL string) *ExchangeService {
	return &ExchangeService{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *ExchangeService) GetRate(from, to string) (float64, error) {
	url := fmt.Sprintf("%s/convert?from=%s&to=%s", s.BaseURL, from, to)
	resp, err := s.Client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()
	var result RateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode error: %w", err)
	}
	fmt.Println(result)
	if resp.StatusCode != http.StatusOK {
		if result.ErrorMsg != "" {
			return 0, fmt.Errorf("api error: %s", result.ErrorMsg)
		}
		return 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return result.Rate, nil
}

// Tests
func TestGetRate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(RateResponse{Base: "USD", Target: "EUR", Rate: 0.92})
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	rate, err := svc.GetRate("USD", "EUR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rate != 0.92 {
		t.Errorf("expected 0.92, got %f", rate)
	}
}

func TestGetRate_APIBusinessError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RateResponse{ErrorMsg: "invalid currency pair"})
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	_, err := svc.GetRate("USD", "XYZ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetRate_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	_, err := svc.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestGetRate_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second) // longer than client timeout
	}))
	defer server.Close()

	svc := &ExchangeService{
		BaseURL: server.URL,
		Client:  &http.Client{Timeout: 100 * time.Millisecond},
	}
	_, err := svc.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestGetRate_ServerPanic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RateResponse{ErrorMsg: "internal error"})
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	_, err := svc.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}

func TestGetRate_EmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// write nothing
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	_, err := svc.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("expected decode error for empty body, got nil")
	}
}
