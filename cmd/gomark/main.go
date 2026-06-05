// Command gomark builds and previews markdown-powered documentation sites.
//
//	gomark build <content-dir> <output-dir>   # render a static site
//	gomark serve <content-dir> [--live]        # preview locally
//
// Production deployments serve the static output of `gomark build` from any
// static host. `gomark serve` is a development tool only.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	gm "github.com/arivictor/gomark"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "build":
		if err := runBuild(os.Args[2:]); err != nil {
			log.Fatalf("gomark build: %v", err)
		}
	case "serve":
		if err := runServe(os.Args[2:]); err != nil {
			log.Fatalf("gomark serve: %v", err)
		}
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "gomark: unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func runBuild(args []string) error {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	title := fs.String("title", "", "site title")
	siteURL := fs.String("url", "", "public site URL, used for canonical links, sitemap, and SEO metadata")
	noRunner := fs.Bool("no-runner", false, "disable the in-browser Go runner")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: gomark build <content-dir> <output-dir> [flags]")
		fs.PrintDefaults()
	}

	positionals, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	if len(positionals) < 2 {
		fs.Usage()
		return fmt.Errorf("expected a content dir and an output dir")
	}

	content, output := positionals[0], positionals[1]
	opts := append(commonOptions(content, *title, *siteURL, *noRunner), gm.WithSiteMode(gm.PreRender))

	if err := gm.NewSite(opts...).Export(output); err != nil {
		return err
	}
	log.Printf("built static site: %s -> %s", content, output)
	if *siteURL == "" {
		log.Printf("tip: pass --url https://your.site for correct canonical links and SEO metadata")
	}
	return nil
}

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	live := fs.Bool("live", false, "render pages on every request and auto-reload the browser when files change")
	port := fs.String("port", "8080", "port to listen on")
	title := fs.String("title", "", "site title")
	siteURL := fs.String("url", "", "public site URL, used for canonical links and SEO metadata")
	noRunner := fs.Bool("no-runner", false, "disable the in-browser Go runner")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: gomark serve <content-dir> [flags]")
		fs.PrintDefaults()
	}

	positionals, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	content := "."
	if len(positionals) > 0 {
		content = positionals[0]
	}

	mode := gm.PreRender
	if *live {
		mode = gm.LiveRender
	}
	opts := append(commonOptions(content, *title, *siteURL, *noRunner), gm.WithSiteMode(mode))

	addr := ":" + *port
	if *live {
		log.Printf("gomark dev server (live) on http://localhost:%s — watching %s for changes, auto-reload enabled", *port, content)
	} else {
		log.Printf("gomark dev server on http://localhost:%s (run with --live to auto-reload on edits)", *port)
	}
	return gm.NewSite(opts...).Serve(addr, *live)
}

func commonOptions(content, title, siteURL string, noRunner bool) []gm.SiteOption {
	opts := []gm.SiteOption{gm.WithSiteContentDir(content)}
	if title != "" {
		opts = append(opts, gm.WithSiteTitle(title))
	}
	if siteURL != "" {
		opts = append(opts, gm.WithSiteURL(siteURL))
	}
	if noRunner {
		opts = append(opts, gm.WithSiteRunnerEnabled(false))
	}
	return opts
}

// parseArgs lets flags appear before or after the positional arguments, so both
// `gomark serve --live ./docs` and `gomark serve ./docs --live` work. It parses
// leading flags, takes one positional, and repeats until the args are consumed.
func parseArgs(fs *flag.FlagSet, args []string) ([]string, error) {
	var positionals []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		args = fs.Args()
		if len(args) == 0 {
			break
		}
		positionals = append(positionals, args[0])
		args = args[1:]
	}
	return positionals, nil
}

func usage() {
	fmt.Fprint(os.Stderr, `gomark — build and preview markdown documentation sites

Usage:
  gomark build <content-dir> <output-dir> [flags]   Render a static site
  gomark serve <content-dir> [flags]                Preview locally

Build flags:
  --title       site title
  --url         public site URL (canonical links, sitemap, SEO)
  --no-runner   disable the in-browser Go runner

Serve flags:
  --live        render live and auto-reload the browser on file changes
  --port        port to listen on (default 8080)
  --title, --url, --no-runner   as above

Examples:
  gomark serve ./my_docs --live
  gomark build ./my_docs ./dist --url https://docs.example.com

The static output of `+"`gomark build`"+` runs on any static host (GitHub Pages,
Netlify, S3, nginx). The in-browser Go runner needs no server. `+"`gomark serve`"+`
is a development tool only.
`)
}
