package sitemap

import (
	"encoding/xml"
	"sort"
	"strings"
)

type urlset struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []url    `xml:"url"`
}
type url struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod"`
	ChangeFreq string `xml:"changefreq,omitempty"`
}

// GenerateSitemap takes base, routes, and content metas, sorts them, and generates the XML
func GenerateSitemap(base string, routes []RouteMeta, content []ContentMeta) string {
	var entries []url

	// 1. Home
	for _, r := range routes {
		if r.URL == "/" {
			entries = append(entries, url{
				Loc:        strings.TrimRight(base, "/") + r.URL,
				LastMod:    r.LastMod,
				ChangeFreq: r.ChangeFreq,
			})
		}
	}
	// 2. Main pages
	mainPages := []string{"/blog", "/about", "/contact"}
	for _, main := range mainPages {
		for _, r := range routes {
			if r.URL == main {
				entries = append(entries, url{
					Loc:        strings.TrimRight(base, "/") + r.URL,
					LastMod:    r.LastMod,
					ChangeFreq: r.ChangeFreq,
				})
			}
		}
	}
	// 3. Blog articles (sorted)
	var blogArticles []url
	for _, c := range content {
		cf := c.ChangeFreq
		if cf == "" {
			cf = "never"
		}
		blogArticles = append(blogArticles, url{
			Loc:        strings.TrimRight(base, "/") + c.URL,
			LastMod:    c.LastMod,
			ChangeFreq: cf,
		})
	}
	sort.Slice(blogArticles, func(i, j int) bool {
		return blogArticles[i].Loc < blogArticles[j].Loc
	})
	entries = append(entries, blogArticles...)
	// 4. Secondary pages (sorted, not in mainPages)
	var secondary []url
	for _, r := range routes {
		if r.URL != "/" && !contains(mainPages, r.URL) && !strings.HasPrefix(r.URL, "/blog/") {
			secondary = append(secondary, url{
				Loc:        strings.TrimRight(base, "/") + r.URL,
				LastMod:    r.LastMod,
				ChangeFreq: r.ChangeFreq,
			})
		}
	}
	sort.Slice(secondary, func(i, j int) bool {
		return secondary[i].Loc < secondary[j].Loc
	})
	entries = append(entries, secondary...)
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
