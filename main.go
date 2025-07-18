package main

import (
	"fmt"
	"gositemap/sitemap"
	"net/url"
	"os"
	"strings"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
)

func main() {
	opts := ParseCLI()

	routesDir := "src/routes"
	blacklistPath := "blacklist.toml"
	outputPath := "static/sitemap.xml"

	if _, err := os.Stat("gositemap.toml"); os.IsNotExist(err) {
		fmt.Print(Yellow + "Config file 'gositemap.toml' not found. Please enter your website base URL (e.g. https://mysite.com): " + Reset)
		var url string
		fmt.Scanln(&url)
		f, ferr := os.Create("gositemap.toml")
		if ferr != nil {
			fmt.Println(Red + "Could not create gositemap.toml: " + ferr.Error() + Reset)
			os.Exit(1)
		}
		f.WriteString("base_url = \"" + url + "\"\n\nblog_content_dir = \"src/lib/content\"\n\nexclude = [\n  \"/admin\",\n  \"/secret\"\n]\n")
		f.Close()
		fmt.Println(Green + "Created gositemap.toml with your base URL." + Reset)
	}

	cfg, _ := sitemap.LoadConfig("gositemap.toml")
	base := "http://localhost"
	if cfg != nil && cfg.BaseURL != "" {
		base = cfg.BaseURL
	}
	// Validate base_url
	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		fmt.Println(Red + "Invalid base_url in config: must be a valid URL (e.g. https://mysite.com)" + Reset)
		os.Exit(1)
	}
	base = strings.TrimRight(base, "/")

	contentTypes := map[string]string{"blog": "src/lib/content"}
	if cfg != nil && len(cfg.ContentTypes) > 0 {
		contentTypes = cfg.ContentTypes
	}
	var allContent []sitemap.ContentMeta
	for slug, dir := range contentTypes {
		freq := "never"
		if cfg != nil && cfg.ChangeFreq != nil && cfg.ChangeFreq[slug] != "" {
			freq = cfg.ChangeFreq[slug]
		}
		metas, _ := sitemap.ScanContent(dir, slug, freq)
		allContent = append(allContent, metas...)
	}

	bl, err := sitemap.LoadBlacklist(blacklistPath)
	if err != nil || bl == nil {
		fmt.Printf(Yellow+"Blacklist ignored (file missing or invalid): %v"+Reset+"\n", err)
		bl = &sitemap.Blacklist{Exclude: []string{}}
	}

	excludeList := []string{}
	if cfg != nil && cfg.Exclude != nil {
		excludeList = cfg.Exclude
	}

	// Pass excludeList to ScanRoutes
	routes, _ := sitemap.ScanRoutes(routesDir, excludeList)

	all := len(routes) + len(allContent)
	if all == 0 {
		if !opts.Quiet {
			fmt.Println(Yellow + "No page or article found, nothing to do." + Reset)
		}
		os.Exit(0)
	}

	for _, r := range routes {
		if !opts.Quiet {
			msg := fmt.Sprintf(Blue+"Detected page: %s (lastmod: %s", r.URL, r.LastMod)
			if r.ChangeFreq != "" {
				msg += ", changefreq: " + r.ChangeFreq
			}
			msg += ")" + Reset
			fmt.Println(msg)
		}
	}
	for _, meta := range allContent {
		if !opts.Quiet {
			fmt.Printf(Blue+"Detected article: %s (lastmod: %s, changefreq: %s)"+Reset+"\n", meta.URL, meta.LastMod, meta.ChangeFreq)
		}
	}

	// Pass changefreq config to sitemap.GenerateSitemap if needed
	xml := sitemap.GenerateSitemap(base, routes, allContent)
	if opts.DryRun {
		if !opts.Quiet {
			fmt.Println(Green + "--- DRY RUN: sitemap.xml output ---" + Reset)
		}
		fmt.Println(xml)
		return
	}
	if err := os.WriteFile(outputPath, []byte(xml), 0644); err != nil {
		fmt.Printf(Red+"Error writing sitemap: %v"+Reset+"\n", err)
		os.Exit(1)
	}
	if !opts.Quiet {
		fmt.Printf(Green+"Sitemap successfully generated (%d entries) in %s"+Reset+"\n", all, outputPath)
	}
}
