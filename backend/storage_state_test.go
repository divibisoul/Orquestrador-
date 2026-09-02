package backend

import (
	"testing"
)

func TestWeb3StorageStatusNotConfiguredWithoutUCAN(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Web3StorageToken = ""
	cfg.Web3StorageURL = ""
	storage := NewWeb3Storage(cfg)
	status := storage.Status()
	if status["status"] != "not-configured" {
		t.Fatalf("expected not-configured status, got %#v", status)
	}
	if status["code"] != storageNotConfiguredCode {
		t.Fatalf("expected %s code, got %#v", storageNotConfiguredCode, status["code"])
	}
	if status["cureSuggestion"] == "" {
		t.Fatal("expected explicit cure suggestion")
	}
}

func TestWeb3StorageUploadDoesNotAttemptWithoutCredential(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Web3StorageToken = ""
	cfg.Web3StorageURL = ""
	storage := NewWeb3Storage(cfg)
	_, _, err := storage.Upload(nil, nil, "ignored")
	if err == nil {
		t.Fatal("expected upload configuration error")
	}
}
