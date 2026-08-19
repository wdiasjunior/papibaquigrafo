package src

import (
  "net/http"
  "net/http/httptest"
  "testing"
  "time"
)

func TestFetchWeebcentral429(t *testing.T) {
  hits := 0
  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    hits++
    if hits <= 2 {
      w.Header().Set("Retry-After", "1")
      w.WriteHeader(429)
      return
    }
    w.Write([]byte("ok"))
  }))
  defer server.Close()

  start := time.Now()
  body, err := fetchWeebcentral(server.URL, "")
  if err != nil {
    t.Fatalf("expected success after 429s, got %v", err)
  }
  if string(body) != "ok" {
    t.Fatalf("wrong body: %q", body)
  }
  if hits != 3 {
    t.Fatalf("expected 3 hits, got %d", hits)
  }
  if elapsed := time.Since(start); elapsed < 2*time.Second {
    t.Fatalf("expected to honor Retry-After (~2s total), took %s", elapsed)
  }
}

func TestFetchWeebcentralHardFailure(t *testing.T) {
  hits := 0
  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    hits++
    w.WriteHeader(404)
  }))
  defer server.Close()

  _, err := fetchWeebcentral(server.URL, "")
  if err == nil {
    t.Fatal("expected error on persistent 404")
  }
  if hits != 3 {
    t.Fatalf("expected exactly 3 attempts, got %d", hits)
  }
}
