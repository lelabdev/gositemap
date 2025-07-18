package sitemap

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ContentMeta struct {
	URL        string
	LastMod    string
	ChangeFreq string
}

// parsePublishDateFromFrontMatter extracts publishDate from YAML frontmatter if present
func parsePublishDateFromFrontMatter(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return time.Now().Format("2006-01-02")
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inFrontMatter := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			if !inFrontMatter {
				inFrontMatter = true
				continue
			} else {
				break
			}
		}
		if inFrontMatter && strings.HasPrefix(line, "publishDate:") {
			date := strings.TrimSpace(strings.TrimPrefix(line, "publishDate:"))
			if len(date) >= 10 {
				return date[:10]
			}
		}
	}
	return time.Now().Format("2006-01-02")
}

// ScanContent returns a slice of ContentMeta (URL + lastmod + changefreq)
func ScanContent(root string, slugPrefix string, changefreq string) ([]ContentMeta, error) {
	var metas []ContentMeta
	entries, err := os.ReadDir(root)
	if err != nil {
		return []ContentMeta{}, nil // If dir does not exist, just return empty
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".svx") {
			slug := strings.TrimSuffix(strings.TrimSuffix(name, ".md"), ".svx")
			url := "/" + slugPrefix + "/" + slug
			url = strings.ReplaceAll(url, "//", "/")
			lastmod := parsePublishDateFromFrontMatter(filepath.Join(root, name))
			metas = append(metas, ContentMeta{URL: url, LastMod: lastmod, ChangeFreq: changefreq})
		}
	}
	return metas, nil
}
