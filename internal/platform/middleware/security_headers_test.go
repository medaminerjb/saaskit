package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler := SecurityHeaders()(nextHandler)

	// Test HTTP request (HSTS should be omitted per RFC 6797)
	req := httptest.NewRequest("GET", "http://localhost/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if rec.Header().Get("Strict-Transport-Security") != "" {
		t.Errorf("expected no HSTS header on HTTP request, got %s", rec.Header().Get("Strict-Transport-Security"))
	}

	expectedHeaders := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"X-XSS-Protection":        "0",
		"Content-Security-Policy": "default-src 'self'; frame-ancestors 'none';",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
		"Permissions-Policy":      "geolocation=(), microphone=(), camera=()",
	}

	for header, expected := range expectedHeaders {
		got := rec.Header().Get(header)
		if got != expected {
			t.Errorf("header %s = %q, want %q", header, got, expected)
		}
	}

	// Test HTTPS request (HSTS present)
	reqHTTPS := httptest.NewRequest("GET", "https://localhost/health", nil)
	recHTTPS := httptest.NewRecorder()

	handler.ServeHTTP(recHTTPS, reqHTTPS)

	if recHTTPS.Header().Get("Strict-Transport-Security") != "max-age=31536000; includeSubDomains; preload" {
		t.Errorf("expected HSTS header on HTTPS request, got %q", recHTTPS.Header().Get("Strict-Transport-Security"))
	}
}
