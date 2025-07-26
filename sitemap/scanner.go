package sitemap

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
) // en haut de ton fichier

type RouteMeta struct {
	URL        string
	LastMod    string
	ChangeFreq string
}

// ScanRoutes returns a slice of RouteMeta (URL + lastmod + changefreq)
func ScanRoutes(root string, exclude []string) ([]RouteMeta, error) {
	var metas []RouteMeta

	validExt := []string{".md", ".svx"}
	baseNames := []string{"+page.svelte"}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
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

		name := d.Name()

		hasValidExt := false
		for _, ext := range validExt {
			if strings.HasSuffix(name, ext) {
				hasValidExt = true
				break
			}
		}

		fmt.Println("-> Scanning:", path)
		if !slices.Contains(baseNames, name) && !hasValidExt {
			return nil
		}

		if strings.Contains(rel, string(os.PathSeparator)+"[") {
			return nil
		}

		// Build URL
		url := "/" + strings.ReplaceAll(rel, "\\", "/")

		for _, ext := range validExt {
			if strings.HasSuffix(url, ext) {
				url = strings.TrimSuffix(url, ext)
				break
			}
		}
		url = strings.TrimSuffix(url, "/+page.svelte")

		// Remove "content" or "src/content" prefix to get clean URL
		if strings.HasPrefix(url, "/src/content/") {
			url = strings.TrimPrefix(url, "/src/content")
		} else if strings.HasPrefix(url, "/content/") {
			url = strings.TrimPrefix(url, "/content")
		}
		if url == "" {
			url = "/"
		}

		if url == "/+page.svelte" || url == "/." {
			url = "/"
		}

		// Exclusion
		for _, ex := range exclude {
			if strings.HasPrefix(ex, "/") {
				if strings.HasPrefix(url, ex) {
					return nil
				}
			} else {
				parts := strings.Split(url, "/")
				if slices.Contains(parts, ex) {
					return nil
				}
			}
		}

		// Last modified
		fi, err := os.Stat(path)
		lastmod := time.Now().Format("2006-01-02")
		if err == nil {
			lastmod = fi.ModTime().Format("2006-01-02")
		}

		// Change frequency
		changefreq := "never"
		if url == "/" {
			changefreq = ""
		} else if url == "/blog" {
			changefreq = "weekly"
		}

		// Clean (flow) segments
		parts := strings.Split(url, "/")
		var clean []string
		for _, part := range parts {
			if part == "" || (strings.HasPrefix(part, "(") && strings.HasSuffix(part, ")")) {
				continue
			}
			clean = append(clean, part)
		}
		url = "/" + strings.Join(clean, "/")

		metas = append(metas, RouteMeta{
			URL:        url,
			LastMod:    lastmod,
			ChangeFreq: changefreq,
		})

		return nil
	})

	return metas, err
}
