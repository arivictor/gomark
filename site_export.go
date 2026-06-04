package gomark

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Export pre-renders every content page to static HTML under outputDir, copies
// the public assets, and writes sitemap.xml, robots.txt, and search-index.json.
// The result is a self-contained static site that needs no Go server: any static
// host (GitHub Pages, S3, nginx) can serve it, and it works offline.
//
// Because there is no HTTP request, absolute URLs (canonical, Open Graph) come
// from the configured site URL — set WithSiteURL for correct SEO metadata.
func (s *Site) Export(outputDir string) error {
	if s == nil {
		return fmt.Errorf("site is nil")
	}
	outputDir = strings.TrimSpace(outputDir)
	if outputDir == "" {
		return fmt.Errorf("export output dir is empty")
	}

	a := &s.App
	b, err := a.buildSite(true)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	pages := 0
	err = eachContentRoute(b.contentDir, func(slug, route, _ string) error {
		page, getErr := b.provider.Get(slug)
		if getErr != nil {
			return fmt.Errorf("render %s: %w", slug, getErr)
		}

		title := page.Title
		if strings.TrimSpace(title) == "" {
			title = pageTitleFromSlug(slug)
		}
		navTitle, nav := b.index.Sidebar(route, b.sidebarDepth)

		data := PageData{
			Title:           title,
			Description:     page.Description,
			SiteName:        b.siteName,
			LogoURL:         b.logoURL,
			CanonicalURL:    joinAbsoluteURL(b.siteURL, route),
			OGImageURL:      joinAbsoluteURL(b.siteURL, b.ogImagePath),
			TwitterImageURL: joinAbsoluteURL(b.siteURL, b.twitterImagePath),
			RunnerEnabled:   b.runnerEnabled,
			Robots:          "index,follow",
			Time:            time.Now().UTC().Format(time.RFC3339),
			MarkdownFile:    page.Path,
			BodyHTML:        template.HTML(page.HTML),
			Headings:        page.Headings,
			HideTOC:         page.HideTOC,
			NavTitle:        navTitle,
			Nav:             nav,
			TopNav:          b.topNav,
			CurrentPath:     route,
			StaticBuild:     true,
		}

		if err := writePageFile(b.renderer, exportFilePath(outputDir, route), data); err != nil {
			return err
		}
		pages++
		return nil
	})
	if err != nil {
		return err
	}
	if pages == 0 {
		return fmt.Errorf("no markdown pages found in %s", b.contentDir)
	}

	// Copy public assets (favicons, og images, vendored JS, wasm_exec.js, …).
	if err := copyFS(outputDir, b.publicFS); err != nil {
		return fmt.Errorf("copy public assets: %w", err)
	}

	// The runner module is served gzipped at /runner.wasm by the live server; a
	// static host can't negotiate Content-Encoding, so ship a decompressed copy.
	if b.runnerEnabled {
		if err := writeDecompressedRunner(outputDir, b.publicFS); err != nil {
			return fmt.Errorf("write runner.wasm: %w", err)
		}
	}

	// Generated files win over any bundled copies (e.g. public/robots.txt).
	if err := os.WriteFile(filepath.Join(outputDir, "sitemap.xml"), []byte(b.sitemapXML), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "robots.txt"), []byte(b.robotsTXT), 0o644); err != nil {
		return err
	}

	indexJSON, err := json.Marshal(b.searchIndex.Entries())
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outputDir, "search-index.json"), indexJSON, 0o644)
}

// writePageFile renders one page to file, closing it explicitly so a write/flush
// error surfaced by Close is not silently dropped (as a deferred Close would).
func writePageFile(renderer *FileTemplateRenderer, file string, data PageData) error {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	if err := renderer.RenderTo(f, "markdown", data); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// exportFilePath maps a public route to its static file: "/" -> index.html and
// "/guides/install" -> guides/install/index.html. Directory-style output lets a
// static host serve extensionless URLs without rewrite rules.
func exportFilePath(outputDir, route string) string {
	rel := strings.Trim(route, "/")
	if rel == "" {
		return filepath.Join(outputDir, "index.html")
	}
	return filepath.Join(outputDir, filepath.FromSlash(rel), "index.html")
}

// copyFS copies every file in src into dst, recreating the directory tree.
func copyFS(dst string, src fs.FS) error {
	return fs.WalkDir(src, ".", func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		return copyFSFile(src, p, filepath.Join(dst, filepath.FromSlash(p)))
	})
}

// copyFSFile copies a single file, closing both handles explicitly so closes run
// (and the write-side Close error is returned) even if io.Copy fails.
func copyFSFile(src fs.FS, srcPath, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	in, err := src.Open(srcPath)
	if err != nil {
		return err
	}
	out, err := os.Create(target)
	if err != nil {
		in.Close()
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		in.Close()
		out.Close()
		return err
	}
	if err := in.Close(); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// writeDecompressedRunner gunzips public/runner.wasm.gz into outputDir/runner.wasm.
func writeDecompressedRunner(outputDir string, src fs.FS) error {
	data, err := fs.ReadFile(src, "runner.wasm.gz")
	if err != nil {
		return err
	}
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer gr.Close()

	out, err := os.Create(filepath.Join(outputDir, "runner.wasm"))
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, gr)
	return err
}
