package agentarsenal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyCatalogAndResolve(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/catalog" {
			_, _ = w.Write([]byte(`{"total":1,"items":[{"source":"superpowers","kind":"skill","path":"skills/verification-before-completion/SKILL.md"}]}`))
			return
		}
		if r.URL.Path == "/v1/resolve" {
			var in map[string]string
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				t.Fatal(err)
			}
			_, _ = w.Write([]byte(`{"source":"` + in["source"] + `","path":"` + in["path"] + `","content":"ok"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	p := &Proxy{baseURL: srv.URL, client: srv.Client()}
	catalog, err := p.Catalog(context.Background(), "superpowers", "skill", 10, 0)
	if err != nil { t.Fatal(err) }
	if catalog["total"] != float64(1) { t.Fatalf("unexpected catalog: %#v", catalog) }

	resolved, err := p.Resolve(context.Background(), "superpowers", "skills/verification-before-completion/SKILL.md")
	if err != nil { t.Fatal(err) }
	if resolved["content"] != "ok" { t.Fatalf("unexpected resolve: %#v", resolved) }
}
