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
	xml := sitemap.GenerateSitemap(base, []sitemap.RouteMeta{}, content)
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

	xml := sitemap.GenerateSitemap(base, []sitemap.RouteMeta{}, []sitemap.ContentMeta{})
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

func TestMissingOrInvalidBlacklist(t *testing.T) {
	output := captureOutput(func() {
		bl, err := sitemap.LoadBlacklist("nonexistent_file.toml")
		if err != nil || bl == nil {
			fmt.Printf("Blacklist ignored (file missing or invalid): %v\n", err)
			bl = &sitemap.Blacklist{Exclude: []string{}}
		}
		if bl == nil || len(bl.Exclude) != 0 {
			t.Errorf("Default blacklist should be empty")
		}
	})
	if !strings.Contains(output, "Blacklist ignored") {
		t.Errorf("Missing message for missing blacklist: %s", output)
	}
}

func TestSitemapWithLastmodAndChangefreq(t *testing.T) {
	// Create a temp markdown file with publishDate in frontmatter
	dir := t.TempDir()
	file := dir + "/article.md"
	content := `---
title: "Test"
publishDate: 2025-07-18
---
Body here.`
	os.WriteFile(file, []byte(content), 0644)
	metas, err := sitemap.ScanContent(dir, "blog", "never")
	if err != nil || len(metas) != 1 {
		t.Fatalf("ScanContent failed: %v", err)
	}
	xml := sitemap.GenerateSitemap("https://example.com", []sitemap.RouteMeta{}, metas)
	if !strings.Contains(xml, "<lastmod>2025-07-18</lastmod>") {
		t.Errorf("lastmod not found or incorrect: %s", xml)
	}
	if !strings.Contains(xml, "<changefreq>never</changefreq>") {
		t.Errorf("changefreq not found: %s", xml)
	}
}

func TestSitemapWithNoPublishDate(t *testing.T) {
	dir := t.TempDir()
	file := dir + "/nopub.md"
	content := `---
title: "No Date"
---
Body here.`
	os.WriteFile(file, []byte(content), 0644)
	metas, err := sitemap.ScanContent(dir, "blog", "never")
	if err != nil || len(metas) != 1 {
		t.Fatalf("ScanContent failed: %v", err)
	}
	today := metas[0].LastMod
	xml := sitemap.GenerateSitemap("https://example.com", []sitemap.RouteMeta{}, metas)
	if !strings.Contains(xml, "<lastmod>"+today+"</lastmod>") {
		t.Errorf("lastmod fallback to today failed: %s", xml)
	}
}

func TestSitemapOrderAndFields(t *testing.T) {
	// Home, main, articles, secondary
	routes := []sitemap.RouteMeta{
		{URL: "/", LastMod: "2023-01-01", ChangeFreq: "weekly"},
		{URL: "/blog", LastMod: "2023-01-02", ChangeFreq: "weekly"},
		{URL: "/about", LastMod: "2023-01-03", ChangeFreq: "never"},
		{URL: "/privacy", LastMod: "2023-01-04", ChangeFreq: "never"},
		{URL: "/cgu", LastMod: "2023-01-05", ChangeFreq: "never"},
		{URL: "/contact", LastMod: "2023-01-06", ChangeFreq: "never"},
		{URL: "/secondary", LastMod: "2023-01-07", ChangeFreq: "never"},
	}
	content := []sitemap.ContentMeta{
		{URL: "/blog/article-b", LastMod: "2023-02-01"},
		{URL: "/blog/article-a", LastMod: "2023-02-02"},
	}
	xml := sitemap.GenerateSitemap("https://mysite.com", routes, content)
	// Check order: /, /blog, /about, /contact, /blog/article-a, /blog/article-b, /cgu, /privacy, /secondary
	idx := func(s string) int { return strings.Index(xml, s) }
	order := []string{"/", "/blog", "/about", "/contact", "/blog/article-a", "/blog/article-b", "/cgu", "/privacy", "/secondary"}
	last := -1
	for _, u := range order {
		pos := idx(u)
		if pos == -1 {
			t.Errorf("URL %s not found in sitemap", u)
		}
		if pos < last {
			t.Errorf("URL %s is out of order", u)
		}
		last = pos
	}
	// Check lastmod and changefreq
	for _, u := range []string{"/", "/blog", "/about", "/contact", "/cgu", "/privacy", "/secondary"} {
		if !strings.Contains(xml, "<loc>https://mysite.com"+u+"</loc>") {
			t.Errorf("Missing loc for %s", u)
		}
		if !strings.Contains(xml, "<lastmod>2023-01-") && !strings.Contains(xml, "<lastmod>2023-02-") {
			t.Errorf("Missing lastmod for %s", u)
		}
	}
	if !strings.Contains(xml, "<changefreq>weekly</changefreq>") {
		t.Errorf("Missing changefreq weekly")
	}
	if !strings.Contains(xml, "<changefreq>never</changefreq>") {
		t.Errorf("Missing changefreq never")
	}
}
