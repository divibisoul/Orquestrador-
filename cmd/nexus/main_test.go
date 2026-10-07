package main

import (
	"net/http/httptest"
	"testing"
)

func TestRequireAppBearer(t *testing.T) {
	t.Setenv("N07_APP_TOKEN", "test-token")
	request := httptest.NewRequest("POST", "/execute", nil)
	if err := requireAppBearer(request); err == nil {
		t.Fatal("expected missing bearer token to fail")
	}

	request.Header.Set("Authorization", "Bearer wrong")
	if err := requireAppBearer(request); err == nil {
		t.Fatal("expected invalid bearer token to fail")
	}

	request.Header.Set("Authorization", "Bearer test-token")
	if err := requireAppBearer(request); err != nil {
		t.Fatalf("expected valid bearer token: %v", err)
	}
}

func TestRequireAppBearerRequiresConfiguration(t *testing.T) {
	t.Setenv("N07_APP_TOKEN", "")
	request := httptest.NewRequest("POST", "/execute", nil)
	if err := requireAppBearer(request); err == nil {
		t.Fatal("expected missing app token configuration to fail closed")
	}
}
