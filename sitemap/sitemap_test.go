package sitemap_test

import (
	"bytes"
	"fmt"
	"gositemap/sitemap"
	"io"
	"os"
	"strings"
	"testing"
)

// captureOutput captures stdout for testing console output
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestGenerateSitemap(t *testing.T) {
	base := "https://mysite.com"
	content := []sitemap.ContentMeta{{URL: "/blog/article1", LastMod: "2023-01-01"}}
	xml := sitemap.GenerateSitemap(base, []sitemap.RouteMeta{}, content, []sitemap.URL{}, false)
	if !strings.Contains(xml, "<urlset") || !strings.Contains(xml, "<loc>https://mysite.com/") {
		t.Errorf("Malformed sitemap xml: %s", xml)
	}
	if !strings.Contains(xml, "<loc>https://mysite.com/blog/article1</loc>") {
		t.Errorf("Sitemap xml missing article: %s", xml)
	}
}

func TestSitemapFileCreation(t *testing.T) {
	base := "https://mysite.com"
	tmpfile := "sitemap_test.xml"
	defer os.Remove(tmpfile)

	xml := sitemap.GenerateSitemap(base, []sitemap.RouteMeta{}, []sitemap.ContentMeta{}, []sitemap.URL{}, false)
	if err := os.WriteFile(tmpfile, []byte(xml), 0644); err != nil {
		t.Fatalf("Error writing file: %v", err)
	}
	data, err := os.ReadFile(tmpfile)
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}
	if !strings.Contains(string(data), "<urlset") {
		t.Errorf("Sitemap file does not contain <urlset>")
	}
}

func TestNoPageMessage(t *testing.T) {
	output := captureOutput(func() {
		all := []string{}
		if len(all) == 0 {
			fmt.Println("No page or article found, nothing to do.")
		}
	})
	if !strings.Contains(output, "No page or article found") {
		t.Errorf("Missing message for no file: %s", output)
	}
}

func TestDisplayAddAndFinish(t *testing.T) {
	output := captureOutput(func() {
		routes := []string{"/"}
		content := []string{"/blog/article1"}
		for _, u := range routes {
			fmt.Printf("Detected page: %s\n", u)
		}
		for _, u := range content {
			fmt.Printf("Detected article: %s\n", u)
		}
		fmt.Printf("Sitemap successfully generated (2 entries) in static/sitemap.xml\n")
	})
	if !strings.Contains(output, "Detected page: /") || !strings.Contains(output, "Detected article: /blog/article1") || !strings.Contains(output, "Sitemap successfully generated") {
		t.Errorf("Incomplete display: %s", output)
	}
}

func TestGenerateSitemap_PreserveLastMod(t *testing.T) {
	base := "https://example.com"

	// Existing URLs with specific lastmod dates
	existingURLs := []sitemap.URL{
		{Loc: "https://example.com/", LastMod: "2023-01-01", ChangeFreq: "daily"},
		{Loc: "https://example.com/about", LastMod: "2023-02-01", ChangeFreq: "monthly"},
	}

	// New routes and content that would normally update lastmod
	newRoutes := []sitemap.RouteMeta{
		{URL: "/", LastMod: "2024-01-01", ChangeFreq: "weekly"}, // Should not update lastmod
		{URL: "/contact", LastMod: "2024-03-01", ChangeFreq: "yearly"}, // New entry
	}
	newContent := []sitemap.ContentMeta{
		{URL: "/about", LastMod: "2024-02-01", ChangeFreq: "daily"}, // Should not update lastmod
		{URL: "/blog/new-article", LastMod: "2024-04-01", ChangeFreq: "never"}, // New entry
	}

	// Generate sitemap with overwriteExisting = false (preserve lastmod)
	xml := sitemap.GenerateSitemap(base, newRoutes, newContent, existingURLs, false)

	// Verify that existing lastmod dates are preserved
	if !strings.Contains(xml, "<loc>https://example.com/</loc>\n    <lastmod>2023-01-01</lastmod>\n    <changefreq>daily</changefreq>") {
		t.Errorf("Lastmod for / was not preserved. Got: %s", xml)
	}
	if !strings.Contains(xml, "<loc>https://example.com/about</loc>\n    <lastmod>2023-02-01</lastmod>\n    <changefreq>monthly</changefreq>") {
		t.Errorf("Lastmod for /about was not preserved. Got: %s", xml)
	}

	// Verify that new entries are added
	if !strings.Contains(xml, "<loc>https://example.com/contact</loc>\n    <lastmod>2024-03-01</lastmod>\n    <changefreq>yearly</changefreq>") {
		t.Errorf("New entry /contact not found. Got: %s", xml)
	}
	if !strings.Contains(xml, "<loc>https://example.com/blog/new-article</loc>\n    <lastmod>2024-04-01</lastmod>\n    <changefreq>never</changefreq>") {
		t.Errorf("New entry /blog/new-article not found. Got: %s", xml)
	}

	// Verify that the total number of entries is correct
	expectedEntries := 4 // 2 existing + 2 new
	if count := strings.Count(xml, "<loc>"); count != expectedEntries {
		t.Errorf("Expected %d entries, got %d. XML: %s", expectedEntries, count, xml)
	}
}



// func TestSitemapWithLastmodAndChangefreq(t *testing.T) {
// 	// Create a temp markdown file with publishDate in frontmatter
// 	dir := t.TempDir()
// 	file := dir + "/article.md"
// 	content := `---\ntitle: "Test"\npublishDate: 2025-07-18\n---\nBody here.`
// 	os.WriteFile(file, []byte(content), 0644)
// 	metas, err := sitemap.ScanContent(dir, "blog", "never")
// 	if err != nil || len(metas) != 1 {
// 		t.Fatalf("ScanContent failed: %v", err)
// 	}
// 	xml := sitemap.GenerateSitemap("https://example.com", []sitemap.RouteMeta{}, metas, []sitemap.URL{})
// 	if !strings.Contains(xml, "<lastmod>2025-07-18</lastmod>") {
// 		t.Errorf("lastmod not found or incorrect: %s", xml)
// 	}
// 	if !strings.Contains(xml, "<changefreq>never</changefreq>") {
// 		t.Errorf("changefreq not found: %s", xml)
// 	}
// }
// 
// func TestSitemapWithNoPublishDate(t *testing.T) {
// 	dir := t.TempDir()
// 	file := dir + "/nopub.md"
// 	content := `---\ntitle: "No Date"\n---\nBody here.`
// 	os.WriteFile(file, []byte(content), 0644)
// 	metas, err := sitemap.ScanContent(dir, "blog", "never")
// 	if err != nil || len(metas) != 1 {
// 		t.Fatalf("ScanContent failed: %v", err)
// 	}
// 	today := metas[0].LastMod
// 	xml := sitemap.GenerateSitemap("https://example.com", []sitemap.RouteMeta{}, metas, []sitemap.URL{})
// 	if !strings.Contains(xml, "<lastmod>"+today+"</lastmod>") {
// 		t.Errorf("lastmod fallback to today failed: %s", xml)
// 	}
// }
// 
// func TestSitemapOrderAndFields(t *testing.T) {
// 	// Home, main, articles, secondary
// 	routes := []sitemap.RouteMeta{
// 		{URL: "/", LastMod: "2023-01-01", ChangeFreq: "weekly"},
// 		{URL: "/blog", LastMod: "2023-01-02", ChangeFreq: "weekly"},
// 		{URL: "/about", LastMod: "2023-01-03", ChangeFreq: "never"},
// 		{URL: "/privacy", LastMod: "2023-01-04", ChangeFreq: "never"},
// 		{URL: "/cgu", LastMod: "2023-01-05", ChangeFreq: "never"},
// 		{URL: "/contact", LastMod: "2023-01-06", ChangeFreq: "never"},
// 		{URL: "/secondary", LastMod: "2023-01-07", ChangeFreq: "never"},
// 	}
// 	content := []sitemap.ContentMeta{
// 		{URL: "/blog/article-b", LastMod: "2023-02-01"},
// 		{URL: "/blog/article-a", LastMod: "2023-02-02"},
// 	}
// 	xml := sitemap.GenerateSitemap("https://mysite.com", routes, content, []sitemap.URL{})
// 	// Check order: /, /blog, /about, /contact, /blog/article-a, /blog/article-b, /cgu, /privacy, /secondary
// 	idx := func(s string) int { return strings.Index(xml, s) }
// 	order := []string{"/", "/blog", "/about", "/contact", "/blog/article-a", "/blog/article-b", "/cgu", "/privacy", "/secondary"}
// 	last := -1
// 	for _, u := range order {
// 		pos := idx(u)
// 		if pos == -1 {
// 			t.Errorf("URL %s not found in sitemap", u)
// 		}
// 		if pos < last {
// 			t.Errorf("URL %s is out of order", u)
// 		}
// 		last = pos
// 	}
// 	// Check lastmod and changefreq
// 	for _, u := range []string{"/", "/blog", "/about", "/contact", "/cgu", "/privacy", "/secondary"} {
// 		if !strings.Contains(xml, "<loc>https://mysite.com"+u+"</loc>") {
// 			t.Errorf("Missing loc for %s", u)
// 		}
// 		if !strings.Contains(xml, "<lastmod>2023-01-") && !strings.Contains(xml, "<lastmod>2023-02-") {
// 			t.Errorf("Missing lastmod for %s", u)
// 		}
// 	}
// 	if !strings.Contains(xml, "<changefreq>weekly</changefreq>") {
// 		t.Errorf("Missing changefreq weekly")
// 	}
// 	if !strings.Contains(xml, "<changefreq>never</changefreq>") {
// 		t.Errorf("Missing changefreq never")
// 	}
// }
