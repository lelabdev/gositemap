package sitemap_test

import (
	"os"
	"path/filepath"
	"testing"

	"gositemap/sitemap"
)

func TestScanRoutes(t *testing.T) {
	t.Run("excludes child routes", func(t *testing.T) {
		// Create a temporary directory structure for testing
		tmpDir, err := os.MkdirTemp("", "test-routes")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		// Create test files and directories
		os.MkdirAll(filepath.Join(tmpDir, "admin", "users"), 0755)
		os.WriteFile(filepath.Join(tmpDir, "admin", "users", "+page.svelte"), []byte("test"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "+page.svelte"), []byte("test"), 0644)

		exclude := []string{"/admin"}
		metas, err := sitemap.ScanRoutes(tmpDir, exclude)
		if err != nil {
			t.Fatalf("ScanRoutes failed: %v", err)
		}

		if len(metas) != 1 {
			t.Errorf("Expected 1 route, got %d", len(metas))
		}

		if metas[0].URL != "/" {
			t.Errorf("Expected route to be '/', got %s", metas[0].URL)
		}
	})

	t.Run("excludes segment", func(t *testing.T) {
		// Create a temporary directory structure for testing
		tmpDir, err := os.MkdirTemp("", "test-routes")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		// Create test files and directories
		os.MkdirAll(filepath.Join(tmpDir, "blog", "(flow)"), 0755)
		os.WriteFile(filepath.Join(tmpDir, "blog", "(flow)", "+page.svelte"), []byte("test"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "+page.svelte"), []byte("test"), 0644)

		exclude := []string{"(flow)"}
		metas, err := sitemap.ScanRoutes(tmpDir, exclude)
		if err != nil {
			t.Fatalf("ScanRoutes failed: %v", err)
		}

		if len(metas) != 1 {
			t.Errorf("Expected 1 route, got %d", len(metas))
		}

		if metas[0].URL != "/" {
			t.Errorf("Expected route to be '/', got %s", metas[0].URL)
		}
	})
}
