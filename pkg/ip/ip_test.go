package ip

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchPublic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("203.0.113.7\n"))
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	got, err := fetchPublic(client, srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if got != "203.0.113.7" {
		t.Fatalf("got %q want 203.0.113.7", got)
	}
}

func TestFetchPublicInvalidBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not-an-ip"))
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	_, err := fetchPublic(client, srv.URL)
	if err == nil {
		t.Fatal("expected error")
	}
}
