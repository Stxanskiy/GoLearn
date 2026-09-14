package auth

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoginLimiter(t *testing.T) {
	l := NewLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !l.Allow("1.2.3.4") {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}
	if l.Allow("1.2.3.4") {
		t.Errorf("4th attempt should be blocked")
	}
	// a different IP is independent
	if !l.Allow("5.6.7.8") {
		t.Errorf("different IP should be allowed")
	}
}

func TestLoginLimiterWindow(t *testing.T) {
	l := NewLimiter(1, 20*time.Millisecond)
	if !l.Allow("ip") {
		t.Fatal("first allowed")
	}
	if l.Allow("ip") {
		t.Fatal("second blocked within window")
	}
	time.Sleep(30 * time.Millisecond)
	if !l.Allow("ip") {
		t.Error("allowed again after window elapsed")
	}
}

func TestClientIP(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.9:5555"
	if got := ClientIP(r); got != "10.0.0.9" {
		t.Errorf("RemoteAddr: want 10.0.0.9, got %s", got)
	}
	r.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	if got := ClientIP(r); got != "203.0.113.5" {
		t.Errorf("XFF: want 203.0.113.5, got %s", got)
	}
}
