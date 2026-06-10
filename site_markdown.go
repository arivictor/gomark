package gomark

import (
	"errors"
	"fmt"
	"html"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// routeFromFrontmatter returns an explicit route override from frontmatter
// (slug/permalink/route), normalized to a clean, fragment-free "/..." path, or
// "" if none is set or it cannot be made safe. Dot-segments are resolved so an
// override can never escape the site root (and, in the exporter, never escape
// the output directory).
func routeFromFrontmatter(meta map[string]string) string {
	if meta == nil {
		return ""
	}
	raw := strings.TrimSpace(firstNonEmpty(meta["slug"], meta["permalink"], meta["route"]))
	if raw == "" {
		return ""
	}
	route := normalizeLinkTarget(raw)
	// Routes are paths, not anchors: drop any fragment.
	if i := strings.IndexByte(route, '#'); i >= 0 {
		route = route[:i]
	}
	if !strings.HasPrefix(route, "/") {
		return ""
	}
	// path.Clean collapses "." and ".." (and ".." at root is dropped), so the
	// result is always a rooted path that stays within the site.
	cleaned := path.Clean(route)
	if !strings.HasPrefix(cleaned, "/") {
		return ""
	}
	return cleaned
}

var (
	ErrInvalidMarkdownPath = errors.New("invalid markdown path")
	ErrMarkdownNotFound    = errors.New("markdown file not found")
)

var orderedListRe = regexp.MustCompile(`^(\d+)\.\s+(.+)$`)
var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

// calloutRe matches a GitHub-style admonition marker (`[!NOTE]`, `[!TIP]`, …) as the
// first line of a blockquote, capturing the kind and any trailing inline text.
var calloutRe = regexp.MustCompile(`(?i)^\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*(.*)$`)

// tableSepCellRe matches a single GFM table delimiter cell (e.g. `---`, `:--`, `--:`,
// `:-:`), used to confirm that a `|`-bearing line is a real table header.
var tableSepCellRe = regexp.MustCompile(`^:?-+:?$`)

// listItem is one collected list entry. depth is the nesting level derived from leading
// indentation (two spaces per level); ordered distinguishes <ol> from <ul>. num is the
// source number of an ordered item (so an interrupted list resumes via <ol start>), and
// task is 0 for a plain item, 1 for an unchecked task ([ ]), or 2 for a checked one ([x]).
type listItem struct {
	depth   int
	ordered bool
	num     int
	task    int
	text    string
}

// Heading is a single in-page heading collected during rendering, used to build
// the on-page table of contents.
type Heading struct {
	Level int
	Text  string
	ID    string
}

type MarkdownRenderer interface {
	Render(markdown string) (html string, headings []Heading)
}

type StdlibMarkdownRenderer struct {
	RunnerEnabled bool
}

type MarkdownService struct {
	renderer   MarkdownRenderer
	contentDir string
}

func NewMarkdownService(renderer MarkdownRenderer, contentDir string) MarkdownService {
	if renderer == nil {
		renderer = StdlibMarkdownRenderer{}
	}
	if strings.TrimSpace(contentDir) == "" {
		contentDir = "content"
	}

	return MarkdownService{
		renderer:   renderer,
		contentDir: filepath.Clean(contentDir),
	}
}

// RenderedPage is the result of loading a markdown file: its resolved path, the
// rendered HTML body, and metadata pulled from optional YAML frontmatter.
type RenderedPage struct {
	Path        string
	HTML        string
	Title       string
	Description string
	Headings    []Heading
	HideTOC     bool
	HideNav     bool
}

func (s MarkdownService) LoadAndRender(slug string) (RenderedPage, error) {
	resolved, err := resolveContentPath(s.contentDir, slug)
	if err != nil {
		return RenderedPage{}, err
	}

	data, readErr := os.ReadFile(resolved)
	if readErr != nil {
		if errors.Is(readErr, os.ErrNotExist) {
			return RenderedPage{}, ErrMarkdownNotFound
		}
		return RenderedPage{}, fmt.Errorf("read markdown %s: %w", resolved, readErr)
	}

	meta, body := parseFrontmatter(string(data))

	title := meta["title"]
	if heading, rest, ok := stripLeadingH1(body); ok {
		body = rest
		if title == "" {
			title = heading
		}
	}

	description := firstNonEmpty(meta["description"], meta["tagline"], meta["lede"])

	html, headings := s.renderer.Render(body)
	headings = limitTOCDepth(headings, tocDepth(meta))

	return RenderedPage{
		Path:        resolved,
		HTML:        html,
		Title:       title,
		Description: description,
		Headings:    headings,
		HideTOC:     tocHidden(meta),
		HideNav:     navHidden(meta),
	}, nil
}

// frontmatterBool interprets a frontmatter value as a boolean, recognizing the
// usual truthy/falsy spellings ("true"/"false", "yes"/"no", "on"/"off",
// "show"/"hide", "1"/"0"). It returns ok=false if the key is absent or the
// value isn't recognized, so callers can fall back to other keys or defaults.
func frontmatterBool(meta map[string]string, key string) (value bool, ok bool) {
	raw, present := meta[key]
	if !present {
		return false, false
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "false", "0", "no", "off", "hide", "none":
		return false, true
	case "true", "1", "yes", "on", "show":
		return true, true
	default:
		return false, false
	}
}

// tocHidden reports whether frontmatter disables the on-page table of contents,
// via either `toc: false` or `show_toc: false` (the latter takes precedence,
// letting a page override a folder/site default explicitly either way).
func tocHidden(meta map[string]string) bool {
	if show, ok := frontmatterBool(meta, "show_toc"); ok {
		return !show
	}
	if show, ok := frontmatterBool(meta, "toc"); ok {
		return !show
	}
	return false
}

// navHidden reports whether frontmatter hides the sidebar navigation for this
// page via `show_nav: false` (also accepts 0/no/off/hide/none).
func navHidden(meta map[string]string) bool {
	if show, ok := frontmatterBool(meta, "show_nav"); ok {
		return !show
	}
	return false
}

// tocDepth returns the maximum heading level to include in the TOC from
// frontmatter `toc_depth` (2 or 3). It defaults to 3 (H2 + H3).
func tocDepth(meta map[string]string) int {
	raw := strings.TrimSpace(firstNonEmpty(meta["toc_depth"], meta["tocdepth"]))
	if raw == "" {
		return 3
	}
	if n, err := strconv.Atoi(raw); err == nil && n >= 2 && n <= 3 {
		return n
	}
	return 3
}

func limitTOCDepth(headings []Heading, maxLevel int) []Heading {
	if maxLevel >= 3 {
		return headings
	}
	out := make([]Heading, 0, len(headings))
	for _, h := range headings {
		if h.Level <= maxLevel {
			out = append(out, h)
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// parseFrontmatter splits an optional leading "---" YAML block from the body.
// It is intentionally minimal (flat key: value pairs) to stay dependency-free;
// values are trimmed of surrounding quotes and keys are lowercased.
func parseFrontmatter(raw string) (map[string]string, string) {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil, raw
	}

	closing := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closing = i
			break
		}
	}
	if closing == -1 {
		return nil, raw
	}

	meta := make(map[string]string)
	for _, line := range lines[1:closing] {
		key, value, found := strings.Cut(line, ":")
		key = strings.ToLower(strings.TrimSpace(key))
		if !found || key == "" {
			continue
		}
		meta[key] = strings.Trim(strings.TrimSpace(value), `"'`)
	}

	body := strings.TrimLeft(strings.Join(lines[closing+1:], "\n"), "\n")
	return meta, body
}

// stripLeadingH1 removes a leading "# Title" line so the body doesn't duplicate
// the page header rendered from frontmatter. It returns the heading text, the
// trimmed body, and whether a leading H1 was found.
func stripLeadingH1(body string) (string, string, bool) {
	lines := strings.Split(body, "\n")
	idx := 0
	for idx < len(lines) && strings.TrimSpace(lines[idx]) == "" {
		idx++
	}
	if idx >= len(lines) {
		return "", body, false
	}

	heading := strings.TrimSpace(lines[idx])
	if !strings.HasPrefix(heading, "# ") {
		return "", body, false
	}

	remaining := append(lines[:idx:idx], lines[idx+1:]...)
	rest := strings.TrimLeft(strings.Join(remaining, "\n"), "\n")
	return strings.TrimSpace(heading[2:]), rest, true
}

func resolveContentPath(contentDir, slug string) (string, error) {
	page := strings.TrimSpace(slug)
	if page == "" {
		page = "home"
	}

	cleanPage := filepath.Clean(page)
	if filepath.IsAbs(cleanPage) {
		return "", ErrInvalidMarkdownPath
	}

	if filepath.Ext(cleanPage) == "" {
		cleanPage += ".md"
	}
	if filepath.Ext(cleanPage) != ".md" {
		return "", ErrInvalidMarkdownPath
	}

	fullPath := filepath.Clean(filepath.Join(contentDir, cleanPage))
	rel, err := filepath.Rel(contentDir, fullPath)
	if err != nil {
		return "", ErrInvalidMarkdownPath
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", ErrInvalidMarkdownPath
	}

	return fullPath, nil
}

func splitFenceInfo(info string) []string {
	info = strings.TrimSpace(info)
	if info == "" {
		return nil
	}

	if prefix, rest, found := strings.Cut(info, ":"); found && !strings.ContainsAny(prefix, " \t\n\r") {
		info = strings.TrimSpace(prefix + " " + rest)
	}

	var tokens []string
	var current strings.Builder
	var quote rune
	justClosedQuote := false

	flush := func() {
		if current.Len() == 0 {
			return
		}
		tokens = append(tokens, current.String())
		current.Reset()
	}

	for _, r := range info {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
				justClosedQuote = true
				continue
			}
			current.WriteRune(r)
		case r == '"' || r == '\'':
			quote = r
		case r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == ':':
			flush()
			justClosedQuote = false
		default:
			if justClosedQuote {
				flush()
				justClosedQuote = false
			}
			current.WriteRune(r)
		}
	}

	flush()
	return tokens
}

func (s StdlibMarkdownRenderer) Render(markdown string) (string, []Heading) {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")

	var out strings.Builder
	var headings []Heading
	seen := map[string]int{}
	makeID := func(text string) string {
		base := slugify(text)
		count := seen[base]
		seen[base]++
		if count == 0 {
			return base
		}
		return fmt.Sprintf("%s-%d", base, count)
	}
	var paragraph []string
	var quote []string
	var listItems []listItem
	var codeLines []string
	inCode := false
	codeLang := ""
	codeTitle := ""
	codeRun := false
	codeEditable := false
	codeGroup := ""
	codeFenceChar := ""
	codeFenceLen := 0

	flushParagraph := func() {
		if len(paragraph) == 0 {
			return
		}
		out.WriteString("<p>")
		out.WriteString(renderInline(strings.Join(paragraph, " ")))
		out.WriteString("</p>\n")
		paragraph = nil
	}

	flushQuote := func() {
		if len(quote) == 0 {
			return
		}
		if m := calloutRe.FindStringSubmatch(quote[0]); m != nil {
			kind := strings.ToLower(m[1])
			body := make([]string, 0, len(quote))
			if rest := strings.TrimSpace(m[2]); rest != "" {
				body = append(body, rest)
			}
			body = append(body, quote[1:]...)
			out.WriteString(`<div class="callout callout-` + kind + `">`)
			out.WriteString(`<p class="callout-title">` + calloutLabel(kind) + `</p>`)
			if len(body) > 0 {
				out.WriteString("<p>")
				out.WriteString(renderInline(strings.Join(body, " ")))
				out.WriteString("</p>")
			}
			out.WriteString("</div>\n")
			quote = nil
			return
		}
		out.WriteString("<blockquote><p>")
		out.WriteString(renderInline(strings.Join(quote, " ")))
		out.WriteString("</p></blockquote>\n")
		quote = nil
	}

	flushList := func() {
		if len(listItems) == 0 {
			return
		}
		for i := 0; i < len(listItems); {
			i = emitList(&out, listItems, i, listItems[0].depth)
		}
		listItems = nil
	}

	flushCode := func() {
		if len(codeLines) == 0 && codeLang == "" {
			return
		}

		title := codeTitle
		if title == "" {
			if codeLang != "" {
				title = codeLang
			} else {
				title = "code"
			}
		}

		classAttr := ""
		if codeLang != "" {
			classAttr = ` class="language-` + html.EscapeString(codeLang) + `"`
		}

		allowRun := s.RunnerEnabled && strings.EqualFold(codeLang, "go") && (codeRun || codeEditable)
		editableAttr := "false"
		if codeEditable {
			editableAttr = "true"
		}

		out.WriteString("<div class=\"code-frame\"")
		if codeEditable {
			out.WriteString(" data-code-editable=\"true\"")
		}
		if allowRun {
			out.WriteString(" data-runner-run=\"true\" data-runner-editable=\"")
			out.WriteString(editableAttr)
			out.WriteString("\"")
		}
		if codeGroup != "" {
			// Adjacent code-frames sharing a group are merged into a tabbed
			// multi-file example by the client; the title becomes the tab label.
			out.WriteString(" data-tab-group=\"")
			out.WriteString(html.EscapeString(codeGroup))
			out.WriteString("\" data-tab-title=\"")
			out.WriteString(html.EscapeString(title))
			out.WriteString("\"")
		}
		out.WriteString(">")
		out.WriteString("<div class=\"code-frame-header\">")
		out.WriteString("<span class=\"code-frame-title\">")
		out.WriteString(html.EscapeString(title))
		out.WriteString("</span>")
		out.WriteString("<div class=\"code-frame-actions\">")
		if allowRun && codeEditable {
			out.WriteString("<button type=\"button\" class=\"code-format\" data-format-code=\"\" aria-label=\"Format code block\" title=\"Format with gofmt\">Format</button>")
		}
		if allowRun {
			out.WriteString("<button type=\"button\" class=\"code-run\" data-run-code=\"\" aria-label=\"Run code block\" title=\"Run in runner\">Run</button>")
		}
		out.WriteString("<button type=\"button\" class=\"code-copy\" data-copy-code=\"\" aria-label=\"Copy code block\" title=\"Copy code\">")
		out.WriteString("<svg viewBox=\"0 0 24 24\" aria-hidden=\"true\" focusable=\"false\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><rect x=\"9\" y=\"9\" width=\"13\" height=\"13\" rx=\"2\" ry=\"2\"></rect><path d=\"M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1\"></path></svg>")
		out.WriteString("</button>")
		out.WriteString("</div>")
		out.WriteString("</div>")
		out.WriteString("<pre><code")
		out.WriteString(classAttr)
		out.WriteString(">")
		out.WriteString(html.EscapeString(strings.Join(codeLines, "\n")))
		out.WriteString("</code></pre>")
		out.WriteString("</div>\n")

		codeLines = nil
		codeLang = ""
		codeTitle = ""
		codeRun = false
		codeEditable = false
		codeGroup = ""
		codeFenceChar = ""
		codeFenceLen = 0
	}

	parseFence := func(line string) (string, int, string, bool) {
		if line == "" {
			return "", 0, "", false
		}
		var fenceChar byte
		switch line[0] {
		case '`', '~':
			fenceChar = line[0]
		default:
			return "", 0, "", false
		}

		count := 0
		for count < len(line) && line[count] == fenceChar {
			count++
		}
		if count < 3 {
			return "", 0, "", false
		}

		return string(fenceChar), count, strings.TrimSpace(line[count:]), true
	}

	parseFenceInfo := func(info string) (string, string, bool, bool, string) {
		tokens := splitFenceInfo(info)
		lang := ""
		title := ""
		run := false
		editable := false
		group := ""

		for _, token := range tokens {
			key, value, found := strings.Cut(token, "=")
			if found {
				switch strings.ToLower(strings.TrimSpace(key)) {
				case "title", "name":
					title = strings.TrimSpace(value)
				case "run":
					run = parseBoolMeta(value)
				case "editable":
					editable = parseBoolMeta(value)
				case "group", "tab":
					group = strings.TrimSpace(value)
				}
				continue
			}

			if lang == "" {
				lang = token
			}
		}

		if title == "" {
			title = lang
		}

		return lang, title, run, editable, group
	}

	for i := 0; i < len(lines); i++ {
		raw := lines[i]
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		depth := (len(line) - len(strings.TrimLeft(line, " "))) / 2

		if inCode {
			if fenceChar, fenceLen, _, ok := parseFence(trimmed); ok && fenceChar == codeFenceChar && fenceLen >= codeFenceLen {
				inCode = false
				flushCode()
				continue
			}
			codeLines = append(codeLines, line)
			continue
		}

		if fenceChar, fenceLen, info, ok := parseFence(trimmed); ok {
			flushParagraph()
			flushQuote()
			flushList()
			inCode = true
			codeFenceChar = fenceChar
			codeFenceLen = fenceLen
			codeLang, codeTitle, codeRun, codeEditable, codeGroup = parseFenceInfo(info)
			codeLines = nil
			continue
		}

		if trimmed == "" {
			flushParagraph()
			flushQuote()
			flushList()
			continue
		}

		if trimmed == "---" {
			flushParagraph()
			flushQuote()
			flushList()
			out.WriteString("<hr />\n")
			continue
		}

		if strings.HasPrefix(trimmed, ">") {
			flushParagraph()
			flushList()
			content := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
			quote = append(quote, content)
			continue
		}

		// GFM table: a `|`-bearing line immediately followed by a delimiter row.
		if strings.Contains(trimmed, "|") && i+1 < len(lines) {
			nextTrim := strings.TrimSpace(strings.TrimRight(lines[i+1], "\r"))
			if aligns, ok := parseTableSeparator(nextTrim); ok {
				flushParagraph()
				flushQuote()
				flushList()

				header := splitTableRow(trimmed)
				out.WriteString("<table>\n<thead>\n<tr>")
				for c, cell := range header {
					al := ""
					if c < len(aligns) {
						al = aligns[c]
					}
					out.WriteString("<th")
					if al != "" {
						out.WriteString(` style="text-align:` + al + `"`)
					}
					out.WriteString(">")
					out.WriteString(renderInline(cell))
					out.WriteString("</th>")
				}
				out.WriteString("</tr>\n</thead>\n<tbody>\n")

				j := i + 2
				for j < len(lines) {
					rowTrim := strings.TrimSpace(strings.TrimRight(lines[j], "\r"))
					if rowTrim == "" || !strings.Contains(rowTrim, "|") {
						break
					}
					cells := splitTableRow(rowTrim)
					out.WriteString("<tr>")
					for c := 0; c < len(header); c++ {
						val := ""
						if c < len(cells) {
							val = cells[c]
						}
						al := ""
						if c < len(aligns) {
							al = aligns[c]
						}
						out.WriteString("<td")
						if al != "" {
							out.WriteString(` style="text-align:` + al + `"`)
						}
						out.WriteString(">")
						out.WriteString(renderInline(val))
						out.WriteString("</td>")
					}
					out.WriteString("</tr>\n")
					j++
				}
				out.WriteString("</tbody>\n</table>\n")
				i = j - 1
				continue
			}
		}

		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			flushParagraph()
			flushQuote()
			task, text := parseTaskListItem(strings.TrimSpace(trimmed[2:]))
			listItems = append(listItems, listItem{depth: depth, ordered: false, task: task, text: text})
			continue
		}

		if matches := orderedListRe.FindStringSubmatch(trimmed); len(matches) == 3 {
			flushParagraph()
			flushQuote()
			num, _ := strconv.Atoi(matches[1])
			listItems = append(listItems, listItem{depth: depth, ordered: true, num: num, text: strings.TrimSpace(matches[2])})
			continue
		}

		headingLevel, headingText, ok := parseHeading(trimmed)
		if ok {
			flushParagraph()
			flushQuote()
			flushList()
			id := makeID(headingText)
			out.WriteString(fmt.Sprintf("<h%d id=\"%s\">%s<a class=\"heading-anchor\" href=\"#%s\" aria-label=\"Permalink to this section\">#</a></h%d>\n", headingLevel, id, renderInline(headingText), id, headingLevel))
			if headingLevel == 2 || headingLevel == 3 {
				headings = append(headings, Heading{Level: headingLevel, Text: headingPlain(headingText), ID: id})
			}
			continue
		}

		if len(quote) > 0 {
			flushQuote()
		}
		if len(listItems) > 0 {
			flushList()
		}

		paragraph = append(paragraph, trimmed)
	}

	if inCode {
		flushCode()
	}
	flushParagraph()
	flushQuote()
	flushList()

	return out.String(), headings
}

func parseBoolMeta(value string) bool {
	v := strings.ToLower(strings.Trim(strings.TrimSpace(value), `"'`))
	switch v {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// slugify turns heading text into a URL-fragment id: lowercase, runs of
// non-alphanumerics collapsed to single hyphens, trimmed.
func slugify(text string) string {
	s := slugRe.ReplaceAllString(strings.ToLower(strings.TrimSpace(text)), "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "section"
	}
	return s
}

// headingPlain strips inline markdown markers so a heading reads cleanly as a
// table-of-contents label.
func headingPlain(text string) string {
	r := strings.NewReplacer("`", "", "**", "", "*", "", "_", "")
	return strings.TrimSpace(r.Replace(text))
}

func parseHeading(line string) (int, string, bool) {
	count := 0
	for count < len(line) && line[count] == '#' {
		count++
	}

	if count < 1 || count > 6 {
		return 0, "", false
	}
	if len(line) <= count || line[count] != ' ' {
		return 0, "", false
	}

	return count, strings.TrimSpace(line[count+1:]), true
}

func renderInline(input string) string {
	remaining := input
	var out strings.Builder

	for {
		start := strings.Index(remaining, "`")
		if start == -1 {
			out.WriteString(renderInlineText(remaining))
			break
		}

		before := remaining[:start]
		out.WriteString(renderInlineText(before))

		rest := remaining[start+1:]
		end := strings.Index(rest, "`")
		if end == -1 {
			out.WriteString(html.EscapeString("`" + rest))
			break
		}

		codeText := rest[:end]
		out.WriteString("<code>")
		out.WriteString(html.EscapeString(codeText))
		out.WriteString("</code>")

		remaining = rest[end+1:]
	}

	return out.String()
}

func renderInlineText(input string) string {
	var escapes []rune
	remaining := protectBackslashEscapes(input, &escapes)
	var out strings.Builder
	var links []struct {
		token string
		html  string
	}
	linkIndex := 0

	for len(remaining) > 0 {
		imgAt := strings.Index(remaining, "![")
		wikiAt := strings.Index(remaining, "[[")
		mdAt := strings.Index(remaining, "[")
		urlAt := indexAutolink(remaining)

		next := -1
		for _, p := range []int{imgAt, wikiAt, mdAt, urlAt} {
			if p >= 0 && (next == -1 || p < next) {
				next = p
			}
		}

		if next == -1 {
			out.WriteString(remaining)
			break
		}

		if next > 0 {
			out.WriteString(remaining[:next])
			remaining = remaining[next:]
			continue
		}

		if strings.HasPrefix(remaining, "![") {
			altEnd := strings.Index(remaining[2:], "]")
			if altEnd == -1 || altEnd+3 >= len(remaining) || remaining[altEnd+3] != '(' {
				// Not a complete image: emit the literal "!" and reprocess the "[".
				out.WriteString(remaining[:1])
				remaining = remaining[1:]
				continue
			}
			altEnd += 2 // index of the closing "]" within remaining

			hrefEnd := findMatchingParen(remaining, altEnd+1)
			if hrefEnd == -1 {
				out.WriteString(remaining[:1])
				remaining = remaining[1:]
				continue
			}

			altText := remaining[2:altEnd]
			src := normalizeLinkTarget(strings.TrimSpace(remaining[altEnd+2 : hrefEnd]))
			if src == "" {
				out.WriteString(remaining[:hrefEnd+1])
			} else {
				token := fmt.Sprintf("@@LINK%d@@", linkIndex)
				linkIndex++
				links = append(links, struct {
					token string
					html  string
				}{
					token: token,
					html:  `<img src="` + html.EscapeString(src) + `" alt="` + html.EscapeString(altText) + `" loading="lazy" />`,
				})
				out.WriteString(token)
			}
			remaining = remaining[hrefEnd+1:]
			continue
		}

		if strings.HasPrefix(remaining, "[[") {
			end := strings.Index(remaining[2:], "]]")
			if end == -1 {
				out.WriteString(html.EscapeString("[["))
				remaining = remaining[2:]
				continue
			}

			inner := strings.TrimSpace(remaining[2 : 2+end])
			href, label := parseWikiLink(inner)
			if href == "" {
				out.WriteString(remaining[:end+4])
			} else {
				token := fmt.Sprintf("@@LINK%d@@", linkIndex)
				linkIndex++
				links = append(links, struct {
					token string
					html  string
				}{
					token: token,
					html:  "<a href=\"" + html.EscapeString(href) + "\">" + renderInline(label) + "</a>",
				})
				out.WriteString(token)
			}
			remaining = remaining[end+4:]
			continue
		}

		// Bare URL autolink (GFM): a http(s):// run not introduced by a "[" link.
		if strings.HasPrefix(remaining, "http://") || strings.HasPrefix(remaining, "https://") {
			end := autolinkEnd(remaining)
			url := remaining[:end]
			token := fmt.Sprintf("@@LINK%d@@", linkIndex)
			linkIndex++
			links = append(links, struct {
				token string
				html  string
			}{
				token: token,
				html:  `<a href="` + html.EscapeString(url) + `">` + html.EscapeString(url) + `</a>`,
			})
			out.WriteString(token)
			remaining = remaining[end:]
			continue
		}

		labelEnd := strings.Index(remaining[1:], "]")
		if labelEnd == -1 {
			out.WriteString(html.EscapeString("["))
			remaining = remaining[1:]
			continue
		}

		labelEnd += 1
		if labelEnd+1 >= len(remaining) || remaining[labelEnd+1] != '(' {
			out.WriteString(remaining[:labelEnd+1])
			remaining = remaining[labelEnd+1:]
			continue
		}

		hrefEnd := findMatchingParen(remaining, labelEnd+1)
		if hrefEnd == -1 {
			out.WriteString(remaining[:labelEnd+2])
			remaining = remaining[labelEnd+2:]
			continue
		}

		label := remaining[1:labelEnd]
		href := normalizeLinkTarget(strings.TrimSpace(remaining[labelEnd+2 : hrefEnd]))
		if href == "" {
			out.WriteString(remaining[:hrefEnd+1])
		} else {
			token := fmt.Sprintf("@@LINK%d@@", linkIndex)
			linkIndex++
			links = append(links, struct {
				token string
				html  string
			}{
				token: token,
				html:  "<a href=\"" + html.EscapeString(href) + "\">" + renderInline(label) + "</a>",
			})
			out.WriteString(token)
		}

		remaining = remaining[hrefEnd+1:]
	}

	escaped := html.EscapeString(out.String())
	formatted := applyInlineFormatting(escaped)
	for _, link := range links {
		formatted = strings.ReplaceAll(formatted, link.token, link.html)
	}

	return restoreBackslashEscapes(formatted, escapes)
}

// protectBackslashEscapes replaces each backslash-escaped ASCII-punctuation
// character ("\*", "\_", "\[" …) with an opaque placeholder, recording the literal
// rune in escapes. This runs before link and emphasis parsing so an escaped
// delimiter is treated as plain text; restoreBackslashEscapes swaps the
// placeholders back (HTML-escaped) once formatting is complete.
func protectBackslashEscapes(input string, escapes *[]rune) string {
	if !strings.ContainsRune(input, '\\') {
		return input
	}
	var b strings.Builder
	for i := 0; i < len(input); i++ {
		if input[i] == '\\' && i+1 < len(input) && isASCIIPunct(input[i+1]) {
			fmt.Fprintf(&b, "@@ESC%d@@", len(*escapes))
			*escapes = append(*escapes, rune(input[i+1]))
			i++
			continue
		}
		b.WriteByte(input[i])
	}
	return b.String()
}

func restoreBackslashEscapes(s string, escapes []rune) string {
	for n, r := range escapes {
		s = strings.ReplaceAll(s, fmt.Sprintf("@@ESC%d@@", n), html.EscapeString(string(r)))
	}
	return s
}

// indexAutolink returns the index of the next bare http(s):// URL in s that begins
// at a word boundary, or -1 if there is none. The boundary check keeps it from
// firing inside tokens like "xhttps://…".
func indexAutolink(s string) int {
	from := 0
	for {
		i := strings.Index(s[from:], "://")
		if i == -1 {
			return -1
		}
		i += from
		start := -1
		if strings.HasSuffix(s[:i], "https") {
			start = i - len("https")
		} else if strings.HasSuffix(s[:i], "http") {
			start = i - len("http")
		}
		if start >= 0 && (start == 0 || !isWordByte(s[start-1])) {
			return start
		}
		from = i + 3
	}
}

// autolinkEnd returns the length of the URL run at the start of s, stopping at
// whitespace or a quote/angle bracket and trimming trailing sentence punctuation
// and unbalanced closing parens (matching GFM's autolink boundary behavior).
func autolinkEnd(s string) int {
	i := 0
	for i < len(s) {
		c := s[i]
		if isSpaceByte(c) || c == '<' || c == '>' || c == '"' || c == '\'' || c == '`' {
			break
		}
		i++
	}
	for i > 0 {
		switch c := s[i-1]; {
		case c == '.' || c == ',' || c == ';' || c == ':' || c == '!' || c == '?':
			i--
			continue
		case c == ')' && strings.Count(s[:i], ")") > strings.Count(s[:i], "("):
			i--
			continue
		}
		break
	}
	return i
}

func parseWikiLink(raw string) (string, string) {
	if raw == "" {
		return "", ""
	}

	target := raw
	label := ""
	if pipe := strings.Index(raw, "|"); pipe >= 0 {
		target = strings.TrimSpace(raw[:pipe])
		label = strings.TrimSpace(raw[pipe+1:])
	}

	href := normalizeLinkTarget(target)
	if href == "" {
		return "", ""
	}

	if label == "" {
		label = defaultWikiLabel(target)
	}
	if label == "" {
		label = href
	}

	return href, label
}

func defaultWikiLabel(target string) string {
	base := target
	if hash := strings.Index(base, "#"); hash >= 0 {
		base = base[:hash]
	}
	base = strings.TrimSuffix(base, ".md")
	base = strings.Trim(base, "/")
	if slash := strings.LastIndex(base, "/"); slash >= 0 {
		base = base[slash+1:]
	}
	base = strings.ReplaceAll(base, "-", " ")
	base = strings.ReplaceAll(base, "_", " ")
	return strings.TrimSpace(base)
}

func normalizeLinkTarget(raw string) string {
	href := strings.TrimSpace(raw)
	if href == "" {
		return ""
	}

	if space := strings.IndexAny(href, " \t"); space >= 0 {
		href = href[:space]
	}
	href = strings.Trim(href, "<>")

	if strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}

	if strings.HasPrefix(href, "#") {
		return "#" + slugify(strings.TrimPrefix(href, "#"))
	}

	anchor := ""
	if hash := strings.Index(href, "#"); hash >= 0 {
		anchor = "#" + slugify(href[hash+1:])
		href = href[:hash]
	}

	href = strings.TrimSuffix(href, ".md")
	href = strings.TrimPrefix(href, "./")
	href = strings.ReplaceAll(href, "\\", "/")

	if strings.HasPrefix(href, "/") {
		cleaned := "/" + strings.Trim(strings.TrimSpace(href), "/")
		if cleaned == "/" {
			return cleaned + anchor
		}
		return cleaned + anchor
	}

	trimmed := strings.Trim(strings.TrimSpace(href), "/")
	if trimmed == "" {
		return "/" + anchor
	}

	parts := strings.Split(trimmed, "/")
	for i, part := range parts {
		parts[i] = slugify(part)
	}

	return "/" + strings.Join(parts, "/") + anchor
}

func findMatchingParen(input string, openIdx int) int {
	if openIdx < 0 || openIdx >= len(input) || input[openIdx] != '(' {
		return -1
	}

	depth := 0
	for i := openIdx; i < len(input); i++ {
		switch input[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}

	return -1
}

// applyInlineFormatting turns emphasis and strikethrough delimiters into tags.
// It runs on the already HTML-escaped, link-tokenized text, so the only delimiter
// runes it sees are literal "*", "_" and "~". Double delimiters are processed
// before single ones (so "**" beats "*"), and underscores follow CommonMark's
// intraword rule so identifiers like some_var_name are never emphasized.
func applyInlineFormatting(s string) string {
	s = applyDelim(s, "~~", "<del>", "</del>", false)
	s = applyDelim(s, "**", "<strong>", "</strong>", false)
	s = applyDelim(s, "__", "<strong>", "</strong>", true)
	s = applyDelim(s, "*", "<em>", "</em>", false)
	s = applyDelim(s, "_", "<em>", "</em>", true)
	return s
}

// applyDelim wraps the shortest non-empty span between a matching open/close pair
// of delim in openTag/closeTag. A delimiter only opens when the character after it
// is non-space and only closes when the character before it is non-space; for
// underscore delimiters the outer side must also be a non-word boundary. Unmatched
// or space-flanked delimiters (e.g. the "*" in "2 * 3") are left as literal text.
func applyDelim(s, delim, openTag, closeTag string, underscore bool) string {
	dl := len(delim)
	var b strings.Builder
	i := 0
	for i < len(s) {
		if !strings.HasPrefix(s[i:], delim) || !delimCanOpen(s, i, dl, underscore) {
			b.WriteByte(s[i])
			i++
			continue
		}

		closeIdx := -1
		for j := i + dl; j+dl <= len(s); j++ {
			// j > i+dl guarantees non-empty content between the delimiters, so
			// adjacent runs like "****" stay literal rather than becoming an empty
			// <strong></strong> span (CommonMark requires non-empty content).
			if j > i+dl && strings.HasPrefix(s[j:], delim) && delimCanClose(s, j, dl, underscore) {
				closeIdx = j
				break
			}
		}
		if closeIdx == -1 {
			b.WriteByte(s[i])
			i++
			continue
		}

		b.WriteString(openTag)
		b.WriteString(s[i+dl : closeIdx])
		b.WriteString(closeTag)
		i = closeIdx + dl
	}
	return b.String()
}

func delimCanOpen(s string, i, dl int, underscore bool) bool {
	after := i + dl
	if after >= len(s) || isSpaceByte(s[after]) {
		return false
	}
	if underscore && i > 0 && isWordByte(s[i-1]) {
		return false
	}
	return true
}

func delimCanClose(s string, j, dl int, underscore bool) bool {
	if j == 0 || isSpaceByte(s[j-1]) {
		return false
	}
	if underscore {
		if after := j + dl; after < len(s) && isWordByte(s[after]) {
			return false
		}
	}
	return true
}

func isSpaceByte(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// isASCIIPunct reports whether b is an ASCII punctuation character — the set of
// runes that a backslash may escape per CommonMark.
func isASCIIPunct(b byte) bool {
	switch {
	case b >= '!' && b <= '/':
		return true
	case b >= ':' && b <= '@':
		return true
	case b >= '[' && b <= '`':
		return true
	case b >= '{' && b <= '~':
		return true
	}
	return false
}

// parseTaskListItem detects a GitHub task-list marker ("[ ]", "[x]", "[X]") at the
// start of an unordered list item, returning its state (0 none, 1 unchecked, 2
// checked) and the item text with the marker stripped.
func parseTaskListItem(text string) (int, string) {
	if len(text) < 3 || text[0] != '[' || text[2] != ']' {
		return 0, text
	}
	if len(text) > 3 && text[3] != ' ' {
		return 0, text
	}
	switch text[1] {
	case ' ':
		return 1, strings.TrimSpace(text[3:])
	case 'x', 'X':
		return 2, strings.TrimSpace(text[3:])
	}
	return 0, text
}

// writeListItemOpen writes the opening <li> for an item, rendering a disabled
// checkbox for task-list items.
func writeListItemOpen(out *strings.Builder, item listItem) {
	if item.task == 0 {
		out.WriteString("  <li>")
		return
	}
	out.WriteString(`  <li class="task-list-item"><input type="checkbox" disabled`)
	if item.task == 2 {
		out.WriteString(" checked")
	}
	out.WriteString("> ")
}

// calloutLabel returns the human-readable title for an admonition kind.
func calloutLabel(kind string) string {
	if kind == "" {
		return ""
	}
	return strings.ToUpper(kind[:1]) + strings.ToLower(kind[1:])
}

// splitTableRow splits a single GFM table row into trimmed cell values, dropping one
// optional leading/trailing pipe and honoring backslash-escaped `\|`.
func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")

	var cells []string
	var cur strings.Builder
	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == '\\' && i+1 < len(line) && line[i+1] == '|' {
			cur.WriteByte('|')
			i++
			continue
		}
		if c == '|' {
			cells = append(cells, strings.TrimSpace(cur.String()))
			cur.Reset()
			continue
		}
		cur.WriteByte(c)
	}
	cells = append(cells, strings.TrimSpace(cur.String()))
	return cells
}

// parseTableSeparator reports whether line is a GFM delimiter row and, if so, returns
// the per-column alignment ("left", "right", "center", or "" for none).
func parseTableSeparator(line string) ([]string, bool) {
	line = strings.TrimSpace(line)
	if !strings.Contains(line, "-") || !strings.Contains(line, "|") {
		return nil, false
	}

	cells := splitTableRow(line)
	if len(cells) == 0 {
		return nil, false
	}

	aligns := make([]string, len(cells))
	for i, cell := range cells {
		cell = strings.TrimSpace(cell)
		if !tableSepCellRe.MatchString(cell) {
			return nil, false
		}
		left := strings.HasPrefix(cell, ":")
		right := strings.HasSuffix(cell, ":")
		switch {
		case left && right:
			aligns[i] = "center"
		case right:
			aligns[i] = "right"
		case left:
			aligns[i] = "left"
		default:
			aligns[i] = ""
		}
	}
	return aligns, true
}

// emitList writes one list (and any deeper nested lists) starting at items[start], whose
// nesting level is depth. It returns the index of the first item it did not consume, so
// the caller can resume. Flat lists (all depth 0) render identically to the pre-nesting
// output: one <ul>/<ol> with <li> children, splitting on a same-depth type change.
func emitList(out *strings.Builder, items []listItem, start, depth int) int {
	ordered := items[start].ordered
	tag := "ul"
	if ordered {
		tag = "ol"
	}

	out.WriteString("<")
	out.WriteString(tag)
	// An ordered list that does not start at 1 (e.g. one resumed after an
	// intervening paragraph, or one that explicitly starts at 0) carries its
	// source number through a start attribute so the rendered numbering matches.
	if ordered && items[start].num != 1 {
		fmt.Fprintf(out, " start=\"%d\"", items[start].num)
	}
	out.WriteString(">\n")

	i := start
	for i < len(items) && items[i].depth >= depth {
		if items[i].depth > depth {
			// A deeper item with no parent at this position: nest it directly.
			i = emitList(out, items, i, items[i].depth)
			continue
		}
		if items[i].ordered != ordered {
			// Same depth, different list type: close this list and let the caller
			// open a new one.
			break
		}

		writeListItemOpen(out, items[i])
		out.WriteString(renderInline(items[i].text))
		if i+1 < len(items) && items[i+1].depth > depth {
			i = emitList(out, items, i+1, items[i+1].depth)
			out.WriteString("</li>\n")
			continue
		}
		out.WriteString("</li>\n")
		i++
	}

	out.WriteString("</")
	out.WriteString(tag)
	out.WriteString(">\n")
	return i
}
