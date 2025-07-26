package main

import (
	"fmt"
	"gositemap/sitemap"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
)

func runApp(stdout, stderr io.Writer, args []string) error {
	opts := ParseCLI(args)

	routesDir := "src/routes"
	outputPath := "static/sitemap.xml"

	if _, err := os.Stat("gositemap.toml"); os.IsNotExist(err) {
		fmt.Fprintf(stdout, Yellow+"Config file 'gositemap.toml' not found. Please enter your website base URL (e.g. https://mysite.com): "+Reset)
		var url string
		fmt.Fscanln(os.Stdin, &url)
		f, ferr := os.Create("gositemap.toml")
		if ferr != nil {
			return fmt.Errorf(Red+"Could not create gositemap.toml: %w"+Reset, ferr)
		}
		f.WriteString("base_url = \"" + url + "\"\n\n# You can exclude routes from the sitemap here.\nexclude = [\n  \"/admin\",\n]\n\n# You can define content types that have frontmatter here.\n[content_types]\nblog = \"src/lib/content\"\n")
		f.Close()
		fmt.Fprintf(stdout, Green+"Created gositemap.toml with your base URL."+Reset+"\n")
	}

	cfg, err := sitemap.LoadConfig("gositemap.toml")
	if err != nil {
		return fmt.Errorf(Red+"Could not load gositemap.toml: %w"+Reset, err)
	}

	base := "http://localhost"
	if cfg != nil && cfg.BaseURL != "" {
		base = cfg.BaseURL
	}
	// Validate base_url
	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf(Red + "Invalid base_url in config: must be a valid URL (e.g. https://mysite.com)" + Reset)
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
		metas, err := sitemap.ScanContent(dir, slug, freq)
		if err != nil {
			fmt.Fprintf(stderr, "Error scanning content in %s: %v\n", dir, err)
			continue
		}
		allContent = append(allContent, metas...)
	}

	if cfg != nil {
		for _, glob := range cfg.Glob {
			for _, pattern := range glob.Paths {
				dirs, err := filepath.Glob(pattern)
				if err != nil {
					fmt.Fprintf(stderr, "Error matching glob pattern '%s': %v\n", pattern, err)
					continue
				}
				for _, dir := range dirs {
					addContent(dir, &allContent, cfg)
				}
			}
		}
	}

	excludeList := []string{}
	if cfg != nil && cfg.Exclude != nil {
		excludeList = cfg.Exclude
	}

	// Pass excludeList to ScanRoutes
	routes, err := sitemap.ScanRoutes(routesDir, excludeList)
	if err != nil {
		fmt.Fprintf(stderr, "Error scanning routes in %s: %v\n", routesDir, err)
		return err // Or handle as appropriate
	}

	var existingURLs []sitemap.URL
	if _, err := os.Stat(outputPath); err == nil {
		loadedURLs, loadErr := sitemap.LoadSitemap(outputPath)
		if loadErr != nil {
			fmt.Fprintf(stderr, "Error loading existing sitemap: %v\n", loadErr)
		} else {
			existingURLs = loadedURLs
		}
	}

	all := len(routes) + len(allContent)
	if all == 0 && len(existingURLs) == 0 {
		if !opts.Quiet {
			fmt.Fprintf(stdout, Yellow+"No page or article found, nothing to do."+Reset+"\n")
		}
		return nil
	}

	for _, r := range routes {
		if !opts.Quiet {
			msg := fmt.Sprintf(Blue+"Detected page: %s (lastmod: %s", r.URL, r.LastMod)
			if r.ChangeFreq != "" {
				msg += ", changefreq: " + r.ChangeFreq
			}
			msg += ")" + Reset
			fmt.Fprintf(stdout, msg+"\n")
		}
	}
	for _, meta := range allContent {
		if !opts.Quiet {
			fmt.Fprintf(stdout, Blue+"Detected article: %s (lastmod: %s, changefreq: %s)"+Reset+"", meta.URL, meta.LastMod, meta.ChangeFreq)
		}
	}

	// Determine if we should overwrite existing sitemap entries
	overwriteExisting := false // Default to false (add only, preserve existing lastmod)
	if cfg.PreserveExisting != nil && !*cfg.PreserveExisting { // If preserve_existing is explicitly false
		overwriteExisting = true // Then we overwrite existing entries
	}

	// Pass changefreq config to sitemap.GenerateSitemap if needed
	xml := sitemap.GenerateSitemap(base, routes, allContent, existingURLs, overwriteExisting) // Pass new arg
	if opts.DryRun {
		if !opts.Quiet {
			fmt.Fprintf(stdout, Green+"--- DRY RUN: sitemap.xml output ---\n"+Reset)
		}
		if !overwriteExisting { // If we are in "add only" mode
			if _, err := os.Stat(outputPath); err == nil {
				if !opts.Quiet {
					fmt.Fprintf(stdout, Yellow+"Sitemap file already exists at %s. In dry run, new entries would be added, existing entries would be preserved.\n"+Reset, outputPath)
                }
                fmt.Fprintf(stdout, xml+"\n") // Still print the XML in dry run, but with the correct message
                return nil // This line is restored
            }
        }
        // If overwriteExisting is true, or no existing sitemap, just print the XML
        fmt.Fprintf(stdout, xml+"\n")
        return nil
	}

	// No longer need the `if shouldSkip { ... return nil }` block here
	// The logic for preserving/overwriting is now inside GenerateSitemap

	if err := os.WriteFile(outputPath, []byte(xml), 0644); err != nil {
		return fmt.Errorf(Red+"Error writing sitemap: %w"+Reset, err)
	}
	if !opts.Quiet {
		fmt.Fprintf(stdout, Green+"Sitemap successfully generated (%d entries) in %s"+Reset+"\n", all, outputPath)
	}
	return nil
}

func addContent(dir string, allContent *[]sitemap.ContentMeta, cfg *sitemap.Config) {
	fi, err := os.Stat(dir)
	if err != nil || !fi.IsDir() {
		return
	}

	slug := filepath.Base(dir)
	freq := "never"
	if cfg != nil && cfg.ChangeFreq != nil {
		if f, ok := cfg.ChangeFreq[dir]; ok {
			freq = f
		} else if f, ok := cfg.ChangeFreq[slug]; ok {
			freq = f
		}
	}
	if metas, err := sitemap.ScanContent(dir, slug, freq); err == nil {
		*allContent = append(*allContent, metas...)
	}
}

func main() {
	if err := runApp(os.Stdout, os.Stderr, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error in runApp: %v\n", err)
		os.Exit(1)
	}
}
