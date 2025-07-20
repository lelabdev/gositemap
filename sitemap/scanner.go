package sitemap

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RouteMeta struct {
	URL        string
	LastMod    string
	ChangeFreq string
}

// ScanRoutes returns a slice of RouteMeta (URL + lastmod + changefreq)
func ScanRoutes(root string, exclude []string) ([]RouteMeta, error) {
	var metas []RouteMeta
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), "[") && strings.HasSuffix(d.Name(), "]") {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "+page.svelte" && d.Name() != "+page.md" {
			return nil
		}
		if strings.Contains(rel, string(os.PathSeparator)+"[") {
			return nil
		}
		url := "/" + strings.TrimSuffix(rel, "/+page.svelte")
		url = strings.TrimSuffix(url, "/+page.md")
		url = strings.ReplaceAll(url, "\\", "/")
		if url == "/+page.svelte" || url == "/+page.md" {
			url = "/"
		}
		if url == "/." {
			url = "/"
		}
		for _, ex := range exclude {
			if strings.HasPrefix(ex, "/") {
				if strings.HasPrefix(url, ex) {
					return nil
				}
			} else {
				parts := strings.Split(url, "/")
				for _, part := range parts {
					if part == ex {
						return nil
					}
				}
			}
		}
		// Get lastmod from file mtime
		fi, err := os.Stat(path)
		lastmod := time.Now().Format("2006-01-02")
		if err == nil {
			lastmod = fi.ModTime().Format("2006-01-02")
		}
		// Set changefreq
		changefreq := "never"
		if url == "/" {
			changefreq = ""
		} else if url == "/blog" {
			changefreq = "weekly"
		}
		// Remove segments like (flow) from the URL
		parts := strings.Split(url, "/")
		var clean []string
		for _, part := range parts {
			if part == "" || (strings.HasPrefix(part, "(") && strings.HasSuffix(part, ")")) {
				continue
			}
			clean = append(clean, part)
		}
		url = "/" + strings.Join(clean, "/")
		metas = append(metas, RouteMeta{URL: url, LastMod: lastmod, ChangeFreq: changefreq})
		return nil
	})
	return metas, nil
}
