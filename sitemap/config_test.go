package sitemap_test

import (
	"os"
	"testing"

	"gositemap/sitemap"
)

func TestLoadConfig(t *testing.T) {
	t.Run("load config with multiple excludes", func(t *testing.T) {
		data := `
			base_url = "https://example.com"
			exclude = ["/admin", "/secret"]
		`
		tmpfile, err := os.CreateTemp("", "test.toml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())
		tmpfile.Write([]byte(data))
		tmpfile.Close()

		cfg, err := sitemap.LoadConfig(tmpfile.Name())
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if len(cfg.Exclude) != 2 {
			t.Errorf("Expected 2 excluded routes, got %d", len(cfg.Exclude))
		}
	})

	t.Run("load config with special characters in excludes", func(t *testing.T) {
		data := `
			base_url = "https://example.com"
			exclude = ["/admin/*", "/sitemap.xml"]
		`
		tmpfile, err := os.CreateTemp("", "test.toml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())
		tmpfile.Write([]byte(data))
		tmpfile.Close()

		cfg, err := sitemap.LoadConfig(tmpfile.Name())
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if len(cfg.Exclude) != 2 {
			t.Errorf("Expected 2 excluded routes, got %d", len(cfg.Exclude))
		}

		if cfg.Exclude[0] != "/admin/*" {
			t.Errorf("Expected '/admin/*' as first exclude, got %s", cfg.Exclude[0])
		}
	})

	t.Run("config file does not exist", func(t *testing.T) {
		_, err := sitemap.LoadConfig("nonexistent.toml")
		if err == nil {
			t.Error("Expected error when config file does not exist, got nil")
		}
	})
}
