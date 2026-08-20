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

func TestFetchWeebcentralSlowButSteady(t *testing.T) {
  old := weebcentralStallTimeout
  weebcentralStallTimeout = 200 * time.Millisecond
  defer func() { weebcentralStallTimeout = old }()

  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    flusher := w.(http.Flusher)
    for i := 0; i < 10; i++ {
      w.Write([]byte("x"))
      flusher.Flush()
      time.Sleep(100 * time.Millisecond) // total 1s >> stall timeout, but always progressing
    }
  }))
  defer server.Close()

  body, err := fetchWeebcentral(server.URL, "")
  if err != nil {
    t.Fatalf("slow but progressing transfer must succeed, got %v", err)
  }
  if len(body) != 10 {
    t.Fatalf("expected 10 bytes, got %d", len(body))
  }
}

func TestFetchWeebcentralStalledBody(t *testing.T) {
  old := weebcentralStallTimeout
  weebcentralStallTimeout = 200 * time.Millisecond
  defer func() { weebcentralStallTimeout = old }()

  hits := 0
  block := make(chan struct{})
  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    hits++
    if hits == 1 {
      w.Write([]byte("partial"))
      w.(http.Flusher).Flush()
      <-block // stall the first attempt forever
      return
    }
    w.Write([]byte("complete"))
  }))
  defer server.Close()
  defer close(block) // must run before server.Close, which waits for handlers

  start := time.Now()
  body, err := fetchWeebcentral(server.URL, "")
  if err != nil {
    t.Fatalf("expected retry after stall to succeed, got %v", err)
  }
  if string(body) != "complete" {
    t.Fatalf("wrong body: %q", body)
  }
  if elapsed := time.Since(start); elapsed < 200*time.Millisecond || elapsed > 5*time.Second {
    t.Fatalf("expected one stall abort then quick retry, took %s", elapsed)
  }
}

// pages are never skipped: errors retry until the fetch succeeds
func TestFetchWeebcentralNeverGivesUp(t *testing.T) {
  hits := 0
  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    hits++
    if hits <= 5 {
      w.WriteHeader(404)
      return
    }
    w.Write([]byte("ok"))
  }))
  defer server.Close()

  body, err := fetchWeebcentral(server.URL, "")
  if err != nil {
    t.Fatalf("expected eventual success, got %v", err)
  }
  if string(body) != "ok" {
    t.Fatalf("wrong body: %q", body)
  }
  if hits != 6 {
    t.Fatalf("expected 6 hits, got %d", hits)
  }
}
