package main

import (
	"flag"
	"fmt"
	"os"
)

type CLIOptions struct {
	DryRun bool
	Quiet  bool
	Help   bool
}

func ParseCLI() CLIOptions {
	opts := CLIOptions{}
	flag.BoolVar(&opts.DryRun, "dry-run", false, "Print sitemap to stdout instead of writing to file")
	flag.BoolVar(&opts.Quiet, "quiet", false, "Suppress all output except errors")
	flag.BoolVar(&opts.Help, "help", false, "Show help and exit")
	flag.BoolVar(&opts.Help, "h", false, "Show help and exit (shorthand)")
	flag.Parse()

	if opts.Help {
		printHelp()
		os.Exit(0)
	}

	return opts
}

func printHelp() {
	help := `GoSitemap - SvelteKit static sitemap generator

Usage:
  go run . [options]
  ./gositemap [options]

Options:
  --help, -h     Show this help message and exit
  --dry-run      Print sitemap.xml to stdout instead of writing to file
  --quiet        Suppress all output except errors

If gositemap.toml does not exist, it will be generated interactively.

Example gositemap.toml:

base_url = "https://yoursite.com"

[content_types]
blog = "src/lib/content"
portfolio = "src/lib/portfolio"

[changefreq]
blog = "weekly"
portfolio = "monthly"

exclude = [
  "/admin",
  "/secret"
]
`
	fmt.Println(help)
}
