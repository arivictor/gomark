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
	"strings"

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
	configPath := fs.String("config", "", "path to a gomark.yaml config file (auto-discovered in the content dir by default)")
	title := fs.String("title", "", "site title")
	siteURL := fs.String("url", "", "public site URL, used for canonical links, sitemap, and SEO metadata")
	publicDir := fs.String("public-dir", "", "directory of static assets overlaid on the bundled ones (your favicons, og-image, logos, …)")
	noRunner := fs.Bool("no-runner", false, "disable the in-browser Go runner")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: gomark build [<content-dir> [<output-dir>]] [flags]")
		fs.PrintDefaults()
	}

	positionals, err := parseArgs(fs, args)
	if err != nil {
		return err
	}

	cfg, cfgPath, err := loadConfig(*configPath, positionals)
	if err != nil {
		return err
	}

	content := positional(positionals, 0, cfg.Build.ContentDir, ".")
	output := positional(positionals, 1, cfg.Build.OutputDir, "")
	if output == "" {
		fs.Usage()
		return fmt.Errorf("expected an output dir (positional arg or build.output_dir in %s)", orNone(cfgPath))
	}

	opts := append(buildOptions(fs, cfg, content, *title, *siteURL, *publicDir, *noRunner), gm.WithSiteMode(gm.PreRender))

	if err := gm.NewSite(opts...).Export(output); err != nil {
		return err
	}
	log.Printf("built static site: %s -> %s", content, output)
	if *siteURL == "" && cfg.URL == "" && strings.TrimSpace(os.Getenv("SITE_URL")) == "" {
		log.Printf("tip: set --url (or `url:` in gomark.yaml) for correct canonical links and SEO metadata")
	}
	return nil
}

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	configPath := fs.String("config", "", "path to a gomark.yaml config file (auto-discovered in the content dir by default)")
	live := fs.Bool("live", false, "render pages on every request and auto-reload the browser when files change")
	port := fs.String("port", "8080", "port to listen on")
	title := fs.String("title", "", "site title")
	siteURL := fs.String("url", "", "public site URL, used for canonical links and SEO metadata")
	publicDir := fs.String("public-dir", "", "directory of static assets overlaid on the bundled ones (your favicons, og-image, logos, …)")
	noRunner := fs.Bool("no-runner", false, "disable the in-browser Go runner")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: gomark serve [<content-dir>] [flags]")
		fs.PrintDefaults()
	}

	positionals, err := parseArgs(fs, args)
	if err != nil {
		return err
	}

	cfg, _, err := loadConfig(*configPath, positionals)
	if err != nil {
		return err
	}

	content := positional(positionals, 0, cfg.Build.ContentDir, ".")

	mode := gm.PreRender
	if *live {
		mode = gm.LiveRender
	}
	opts := append(buildOptions(fs, cfg, content, *title, *siteURL, *publicDir, *noRunner), gm.WithSiteMode(mode))

	addr := ":" + *port
	if *live {
		log.Printf("gomark dev server (live) on http://localhost:%s — watching %s for changes, auto-reload enabled", *port, content)
	} else {
		log.Printf("gomark dev server on http://localhost:%s (run with --live to auto-reload on edits)", *port)
	}
	return gm.NewSite(opts...).Serve(addr, *live)
}

// loadConfig loads gomark.yaml: an explicit --config path when given, otherwise
// auto-discovery in the (optional) content-dir positional and the current
// directory. A missing auto-discovered file is not an error.
func loadConfig(configPath string, positionals []string) (*gm.FileConfig, string, error) {
	path := strings.TrimSpace(configPath)
	if path == "" {
		dirs := []string{"."}
		if len(positionals) > 0 && strings.TrimSpace(positionals[0]) != "" {
			dirs = []string{positionals[0], "."}
		}
		path = gm.DiscoverConfigFile(dirs...)
	}
	if path == "" {
		return &gm.FileConfig{}, "", nil
	}

	cfg, err := gm.LoadConfigFile(path)
	if err != nil {
		return nil, path, fmt.Errorf("load config %s: %w", path, err)
	}
	log.Printf("using config %s", path)
	return cfg, path, nil
}

// buildOptions layers configuration sources so that, highest precedence first,
// CLI flags > environment variables > gomark.yaml > defaults. Later options win,
// so the YAML options are appended first, then env, then explicitly-set flags.
func buildOptions(fs *flag.FlagSet, cfg *gm.FileConfig, content, title, siteURL, publicDir string, noRunner bool) []gm.SiteOption {
	opts := cfg.Options()

	if env := strings.TrimSpace(os.Getenv("SITE_URL")); env != "" {
		opts = append(opts, gm.WithSiteURL(env))
	}

	// The content dir always comes from the resolved positional/config value.
	opts = append(opts, gm.WithSiteContentDir(content))

	set := setFlags(fs)
	if set["title"] && title != "" {
		opts = append(opts, gm.WithSiteTitle(title))
	}
	if set["url"] && siteURL != "" {
		opts = append(opts, gm.WithSiteURL(siteURL))
	}
	if set["public-dir"] && publicDir != "" {
		opts = append(opts, gm.WithSitePublicDir(publicDir))
	}
	if set["no-runner"] {
		// GetRunnerEnabled consults PLAYGROUND_ENABLED before the option, so an
		// explicit flag must clear the env var to actually take precedence over it.
		os.Unsetenv("PLAYGROUND_ENABLED")
		opts = append(opts, gm.WithSiteRunnerEnabled(!noRunner))
	}
	return opts
}

// setFlags reports which flags were explicitly provided on the command line.
func setFlags(fs *flag.FlagSet) map[string]bool {
	set := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })
	return set
}

// positional returns the i-th positional argument, falling back to the config
// value and then a default.
func positional(positionals []string, i int, configValue, fallback string) string {
	if i < len(positionals) && strings.TrimSpace(positionals[i]) != "" {
		return positionals[i]
	}
	if strings.TrimSpace(configValue) != "" {
		return configValue
	}
	return fallback
}

func orNone(s string) string {
	if strings.TrimSpace(s) == "" {
		return "gomark.yaml"
	}
	return s
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
  gomark build [<content-dir> [<output-dir>]] [flags]   Render a static site
  gomark serve [<content-dir>] [flags]                  Preview locally

Configuration:
  Site title, logo, SEO (Open Graph / Twitter / description), navigation,
  social links, analytics, and build options can be set in a gomark.yaml file.
  It is auto-discovered in the content dir (and the current directory), or pass
  --config <path>. Resolution precedence, highest first:
    CLI flag  >  environment variable  >  gomark.yaml  >  default

Build flags:
  --config       path to gomark.yaml (auto-discovered by default)
  --title        site title
  --url          public site URL (canonical links, sitemap, SEO)
  --public-dir   static assets overlaid on the bundled ones (favicons, og-image, logos)
  --no-runner    disable the in-browser Go runner

Serve flags:
  --live         render live and auto-reload the browser on file changes
  --port         port to listen on (default 8080)
  --config, --title, --url, --public-dir, --no-runner   as above

Examples:
  gomark serve ./my_docs --live
  gomark build ./my_docs ./dist --url https://docs.example.com
  gomark build                       # paths and options from gomark.yaml

The static output of `+"`gomark build`"+` runs on any static host (GitHub Pages,
Netlify, S3, nginx). The in-browser Go runner needs no server. `+"`gomark serve`"+`
is a development tool only.
`)
}
