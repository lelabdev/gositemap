package sitemap_test

import (
	"gositemap/sitemap"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanContent(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "article1.md"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "article2.md"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "not-md.txt"), []byte(""), 0644)

	urls, err := sitemap.ScanContent(dir, "blog", "never")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"/blog/article1", "/blog/article2"}
	if len(urls) != len(want) {
		t.Fatalf("got %d urls, want %d", len(urls), len(want))
	}
	for i, u := range want {
		if urls[i].URL != u {
			t.Errorf("got %q, want %q", urls[i].URL, u)
		}
	}
}

func TestCustomBlogContentDirFromConfig(t *testing.T) {
	tmpRoot := t.TempDir()
	customDir := filepath.Join(tmpRoot, "myblog")
	os.MkdirAll(customDir, 0755)
	os.WriteFile(filepath.Join(customDir, "foo.md"), []byte(""), 0644)
	cfgPath := filepath.Join(tmpRoot, "gositemap.toml")
	os.WriteFile(cfgPath, []byte(`base_url = "https://mysite.com"
[content_types]
blog = "`+customDir+`"
exclude = []
`), 0644)

	cfg, err := sitemap.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	blogDir := cfg.ContentTypes["blog"]
	urls, err := sitemap.ScanContent(blogDir, "blog", "never")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(urls) != 1 || urls[0].URL != "/blog/foo" {
		t.Errorf("expected /blog/foo, got %+v", urls)
	}
}

func TestScanContentWithSvx(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "article1.svx"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "article2.md"), []byte(""), 0644)

	urls, err := sitemap.ScanContent(dir, "blog", "never")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"/blog/article1", "/blog/article2"}
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

func TestIntegration_SvelteKitProject(t *testing.T) {
	tmpRoot := t.TempDir()
	// Create routes
	routesDir := filepath.Join(tmpRoot, "src", "routes")
	os.MkdirAll(filepath.Join(routesDir, "about"), 0755)
	os.WriteFile(filepath.Join(routesDir, "+page.svelte"), []byte(""), 0644)
	os.WriteFile(filepath.Join(routesDir, "about", "+page.svelte"), []byte(""), 0644)
	os.MkdirAll(filepath.Join(routesDir, "admin"), 0755)
	os.WriteFile(filepath.Join(routesDir, "admin", "+page.svelte"), []byte(""), 0644)
	// Create blog content
	blogDir := filepath.Join(tmpRoot, "src", "lib", "content")
	os.MkdirAll(blogDir, 0755)
	os.WriteFile(filepath.Join(blogDir, "foo.md"), []byte("---\npublishDate: 2023-01-01\n---\n"), 0644)
	os.WriteFile(filepath.Join(blogDir, "bar.svx"), []byte("---\npublishDate: 2023-01-02\n---\n"), 0644)
	// Create TOML config
	toml := `base_url = "https://mysite.com"
[content_types]
blog = "` + blogDir + `"
exclude = ["admin"]
`
	os.WriteFile(filepath.Join(tmpRoot, "gositemap.toml"), []byte(toml), 0644)
	// Load config
	cfg, err := sitemap.LoadConfig(filepath.Join(tmpRoot, "gositemap.toml"))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	routes, _ := sitemap.ScanRoutes(routesDir, cfg.Exclude)
	allContent := []sitemap.ContentMeta{}
	for slug, dir := range cfg.ContentTypes {
		metas, _ := sitemap.ScanContent(dir, slug, "never")
		allContent = append(allContent, metas...)
	}
	xml := sitemap.GenerateSitemap(cfg.BaseURL, routes, allContent)
	if !strings.Contains(xml, "/about") || !strings.Contains(xml, "/blog/foo") || !strings.Contains(xml, "/blog/bar") {
		t.Errorf("sitemap missing expected urls: %s", xml)
	}
	if strings.Contains(xml, "/admin") {
		t.Errorf("sitemap should not contain excluded url /admin: %s", xml)
	}
}
