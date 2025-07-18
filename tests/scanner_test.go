package sitemap_test

import (
	"gositemap/sitemap"
	"os"
	"path/filepath"
	"testing"
)

func TestScanRoutes(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "blog"), 0755)
	os.WriteFile(filepath.Join(dir, "+page.svelte"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "[slug]", "+page.svelte"), []byte(""), 0755)
	os.MkdirAll(filepath.Join(dir, "admin"), 0755)
	os.WriteFile(filepath.Join(dir, "admin", "+page.svelte"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "blog", "+page.md"), []byte(""), 0644)
	os.MkdirAll(filepath.Join(dir, "(flow)", "rendez-vous"), 0755)
	os.MkdirAll(filepath.Join(dir, "(flow)", "rendez-vous", "paiement"), 0755)
	os.WriteFile(filepath.Join(dir, "(flow)", "rendez-vous", "paiement", "+page.svelte"), []byte(""), 0644)

	exclude := []string{"admin"}
	urls, err := sitemap.ScanRoutes(dir, exclude)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"/", "/blog", "/rendez-vous/paiement"}
	if len(urls) != len(want) {
		t.Fatalf("got %d urls, want %d", len(urls), len(want))
	}
	found := make(map[string]bool)
	for _, u := range urls {
		found[u.URL] = true
	}
	for _, u := range want {
		if !found[u] {
			t.Errorf("missing url: %q", u)
		}
	}
}
