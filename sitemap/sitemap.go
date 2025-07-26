package sitemap

import (
	"encoding/xml"
	"os"
	"sort"
	"strings"
)

type urlset struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}
type URL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod"`
	ChangeFreq string `xml:"changefreq,omitempty"`
}

// LoadSitemap reads an XML sitemap file and returns its URLs.
func LoadSitemap(path string) ([]URL, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var us urlset
	if err := xml.Unmarshal(data, &us); err != nil {
		return nil, err
	}
	return us.URLs, nil
}

// GenerateSitemap takes base, routes, content metas, and existing URLs, sorts them, and generates the XML
func GenerateSitemap(base string, routes []RouteMeta, content []ContentMeta, existingURLs []URL, overwriteExisting bool) string {
	uniqueEntries := make(map[string]URL)

	// If not overwriting, add existing URLs to the map first
	if !overwriteExisting {
		for _, u := range existingURLs {
			uniqueEntries[u.Loc] = u
		}
	}

	// Add new routes
	for _, r := range routes {
		loc := strings.TrimRight(base, "/") + r.URL
		if existingURL, ok := uniqueEntries[loc]; ok && overwriteExisting {
			// If overwriteExisting is true, update existing entry with new data
			existingURL.LastMod = r.LastMod
			existingURL.ChangeFreq = r.ChangeFreq
			uniqueEntries[loc] = existingURL
		} else if !ok { // Only add if not already present
			uniqueEntries[loc] = URL{
				Loc:        loc,
				LastMod:    r.LastMod,
				ChangeFreq: r.ChangeFreq,
			}
		}
	}

	// Add new content
	for _, c := range content {
		loc := strings.TrimRight(base, "/") + c.URL
		cf := c.ChangeFreq
		if cf == "" {
			cf = "never"
		}

		if existingURL, ok := uniqueEntries[loc]; ok && overwriteExisting {
			// If overwriteExisting is true, update existing entry with new data
			existingURL.LastMod = c.LastMod
			existingURL.ChangeFreq = cf
			uniqueEntries[loc] = existingURL
		} else if !ok { // Only add if not already present
			uniqueEntries[loc] = URL{
				Loc:        loc,
				LastMod:    c.LastMod,
				ChangeFreq: cf,
			}
		}
	}

	var entries []URL
	for _, u := range uniqueEntries {
		entries = append(entries, u)
	}

	// Sort entries by URL
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Loc < entries[j].Loc
	})

	us := urlset{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  entries,
	}
	out, err := xml.MarshalIndent(us, "", "  ")
	if err != nil {
		return ""
	}
	return xml.Header + string(out)
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}