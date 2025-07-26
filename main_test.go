package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Helper to create a dummy gositemap.toml
func createConfig(t *testing.T, dir string, skipExisting bool) {
	configContent := `base_url = "https://example.com"
output_path = "static/sitemap.xml"
`
	if skipExisting {
		configContent += `preserve_existing = true
`
	} else {
		configContent += `preserve_existing = false
`
	}

	err := os.WriteFile(filepath.Join(dir, "gositemap.toml"), []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create gositemap.toml: %v", err)
	}
}

// Helper to create a dummy sitemap.xml
func createSitemap(t *testing.T, dir string, content string) {
	sitemapDir := filepath.Join(dir, "static")
	err := os.MkdirAll(sitemapDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create static directory: %v", err)
	}
	// Ensure content is a valid sitemap XML
	if !strings.Contains(content, "<urlset") {
		content = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
` + content + `
</urlset>`
	}
	err = os.WriteFile(filepath.Join(sitemapDir, "sitemap.xml"), []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create sitemap.xml: %v", err)
	}
}

func TestRunAppSkipExisting(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer func() {
		os.Chdir(originalWd) // Change back to original directory
	}()
	os.Chdir(tempDir) // Change to temp directory for the test

	// Build the gositemap binary
	binPath := filepath.Join(tempDir, "gositemap-test-bin")
	cmd := exec.Command("go", "build", "-o", binPath, "main.go", "cli.go")
	cmd.Dir = originalWd // Build from the project root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, out)
	}

	// Create dummy src/routes and src/lib/content directories to prevent errors
	os.MkdirAll(filepath.Join(tempDir, "src", "routes"), 0755)
	os.WriteFile(filepath.Join(tempDir, "src", "routes", "+page.svelte"), []byte(""), 0644)
	os.MkdirAll(filepath.Join(tempDir, "src", "lib", "content"), 0755)
	os.WriteFile(filepath.Join(tempDir, "src", "lib", "content", "foo.md"), []byte(""), 0644)


	// Test case 1: preserve_existing = true (default), sitemap exists - should add new and preserve old lastmod
		t.Run("Add new entries and preserve existing lastmod", func(t *testing.T) {
			createConfig(t, tempDir, true)
			initialContent := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://example.com/old-page</loc>
    <lastmod>2020-01-01</lastmod>
  </url>
</urlset>`
			createSitemap(t, tempDir, initialContent)

			cmd := exec.Command(binPath)
			cmd.Dir = tempDir
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Command returned an error: %v\n%s", err, out)
			}

			finalContent, err := os.ReadFile(filepath.Join(tempDir, "static", "sitemap.xml"))
			if err != nil {
				t.Fatalf("Failed to read sitemap.xml after run: %v", err)
			}

			// Expect old entry to be preserved and new entries to be added
			if !strings.Contains(string(finalContent), "<loc>https://example.com/old-page</loc>\n    <lastmod>2020-01-01</lastmod>") {
				t.Errorf("Old entry lastmod not preserved. Got: %q", string(finalContent))
			}
			if !strings.Contains(string(finalContent), "<loc>https://example.com/</loc>") || !strings.Contains(string(finalContent), "<loc>https://example.com/blog/foo</loc>") {
				t.Errorf("New entries not found. Got: %q", string(finalContent))
			}
			if !strings.Contains(string(out), "Sitemap successfully generated") {
				t.Errorf("Expected success message not found in stdout: %s", out)
			}
		})

		// Test case 2: preserve_existing = false, sitemap exists - should overwrite all
		t.Run("Overwrite existing sitemap", func(t *testing.T) {
			// Clean up from previous test run
			os.RemoveAll(filepath.Join(tempDir, "static"))
			os.Remove(filepath.Join(tempDir, "gositemap.toml"))

			createConfig(t, tempDir, false)
			initialContent := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://example.com/old-page</loc>
    <lastmod>2020-01-01</lastmod>
  </url>
</urlset>`
			createSitemap(t, tempDir, initialContent)

			cmd := exec.Command(binPath)
			cmd.Dir = tempDir
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Command returned an error: %v\n%s", err, out)
			}

			finalContent, err := os.ReadFile(filepath.Join(tempDir, "static", "sitemap.xml"))
			if err != nil {
				t.Fatalf("Failed to read sitemap.xml after run: %v", err)
			}

			// Expect old entry to be overwritten with new lastmod
			if strings.Contains(string(finalContent), "<loc>https://example.com/old-page</loc>\n    <lastmod>2020-01-01</lastmod>") {
				t.Errorf("Old entry lastmod was not overwritten. Got: %q", string(finalContent))
			}
			if !strings.Contains(string(finalContent), "<loc>https://example.com/</loc>") || !strings.Contains(string(finalContent), "<loc>https://example.com/blog/foo</loc>") {
				t.Errorf("New entries not found. Got: %q", string(finalContent))
			}
			if !strings.Contains(string(out), "Sitemap successfully generated") {
				t.Errorf("Expected success message not found in stdout: %s", out)
			}
		})

		// Test case 3: preserve_existing = true, sitemap does not exist - should generate
		t.Run("Generate sitemap when not existing and preserve_existing is true", func(t *testing.T) {
			// Clean up from previous test run
			os.RemoveAll(filepath.Join(tempDir, "static"))
			os.Remove(filepath.Join(tempDir, "gositemap.toml"))

			createConfig(t, tempDir, true)
			os.MkdirAll(filepath.Join(tempDir, "static"), 0755)
			cmd := exec.Command(binPath)
			cmd.Dir = tempDir
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Command returned an error: %v\n%s", err, out)
			}

			_, err = os.Stat(filepath.Join(tempDir, "static", "sitemap.xml"))
			if os.IsNotExist(err) {
				t.Errorf("Sitemap was not generated when it should have been.")
			}
			if !strings.Contains(string(out), "Sitemap successfully generated") {
				t.Errorf("Expected success message not found in stdout: %s", out)
			}
		})

		// Test case 4: preserve_existing not set (default behavior), sitemap exists - should add new and preserve old lastmod
		t.Run("Default behavior (preserve_existing not set), sitemap exists", func(t *testing.T) {
			// Clean up from previous test run
			os.RemoveAll(filepath.Join(tempDir, "static"))
			os.Remove(filepath.Join(tempDir, "gositemap.toml"))

			// Create config without preserve_existing
			configContent := `base_url = "https://example.com"
output_path = "static/sitemap.xml"
`
			fmt.Printf("Debug: Writing gositemap.toml with content:\n%s\n", configContent)
			err := os.WriteFile(filepath.Join(tempDir, "gositemap.toml"), []byte(configContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create gositemap.toml: %v", err)
			}

			initialContent := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://example.com/old-default-page</loc>
    <lastmod>2020-01-01</lastmod>
  </url>
</urlset>`
			createSitemap(t, tempDir, initialContent)

			cmd := exec.Command(binPath)
			cmd.Dir = tempDir
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Command returned an error: %v\n%s", err, out)
			}

			finalContent, err := os.ReadFile(filepath.Join(tempDir, "static", "sitemap.xml"))
			if err != nil {
				t.Fatalf("Failed to read sitemap.xml after run: %v", err)
			}

			// Expect old entry to be preserved and new entries to be added
			if !strings.Contains(string(finalContent), "<loc>https://example.com/old-default-page</loc>\n    <lastmod>2020-01-01</lastmod>") {
				t.Errorf("Old default entry lastmod not preserved. Got: %q", string(finalContent))
			}
			if !strings.Contains(string(finalContent), "<loc>https://example.com/</loc>") || !strings.Contains(string(finalContent), "<loc>https://example.com/blog/foo</loc>") {
				t.Errorf("New entries not found in default behavior. Got: %q", string(finalContent))
			}
			if !strings.Contains(string(out), "Sitemap successfully generated") {
				t.Errorf("Expected success message not found in stdout for default behavior: %s", out)
			}
		})
	}
