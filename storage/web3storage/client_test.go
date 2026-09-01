package web3storage

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestUpload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/upload" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("authorization header was not propagated")
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data;") {
			t.Fatalf("expected multipart upload")
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		f, h, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("file field: %v", err)
		}
		defer f.Close()
		if h.Filename != "hello.txt" {
			t.Fatalf("filename=%q", h.Filename)
		}
		data, _ := io.ReadAll(f)
		if string(data) != "hello" {
			t.Fatalf("data=%q", data)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"cid":"bafy-test"}`)
	}))
	defer srv.Close()

	c, err := New(Config{APIBaseURL: srv.URL, Token: "secret", Timeout: time.Second, MaxUpload: 1 << 20})
	if err != nil {
		t.Fatal(err)
	}
	cid, err := c.Upload(context.Background(), "hello.txt", 5, strings.NewReader("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if cid != "bafy-test" {
		t.Fatalf("cid=%q", cid)
	}
}

func TestUploadRejectsOversize(t *testing.T) {
	c, err := New(Config{APIBaseURL: "https://example.invalid", Token: "secret", Timeout: time.Second, MaxUpload: 4})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Upload(context.Background(), "x", 5, strings.NewReader("hello")); err == nil {
		t.Fatal("expected oversize error")
	}
}

func TestUploadRequiresToken(t *testing.T) {
	_, err := New(Config{APIBaseURL: "https://example.invalid", Timeout: time.Second, MaxUpload: 1})
	if err == nil {
		t.Fatal("expected configuration error")
	}
}
