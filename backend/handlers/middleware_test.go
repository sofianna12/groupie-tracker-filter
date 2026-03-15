package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoverMiddleware(t *testing.T) {
	// Handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Wrap with recover middleware
	wrapped := RecoverMiddleware(panicHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Should not panic
	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestRecoverMiddleware_NoPanic(t *testing.T) {
	// Normal handler
	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrapped := RecoverMiddleware(normalHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "OK" {
		t.Errorf("expected body 'OK', got %s", w.Body.String())
	}
}
