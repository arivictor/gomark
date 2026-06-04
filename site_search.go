package gomark

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SearchResult is one matched markdown page returned by the search API.
type SearchResult struct {
	Title   string `json:"title"`
	Path    string `json:"path"`
	Snippet string `json:"snippet,omitempty"`
}

// SearchEntry is one document in the static search index (search-index.json),
// consumed by the client-side search used in exported static builds.
type SearchEntry struct {
	Title string `json:"title"`
	Path  string `json:"path"`
	Body  string `json:"body"`
}

// Entries returns every indexed document for serialization into a static search
// index. The body is the same normalized text the server-side query scores on.
func (idx *SearchIndex) Entries() []SearchEntry {
	if idx == nil {
		return nil
	}
	out := make([]SearchEntry, 0, len(idx.docs))
	for _, doc := range idx.docs {
		out = append(out, SearchEntry{Title: doc.title, Path: doc.path, Body: doc.body})
	}
	return out
}

type searchDoc struct {
	title      string
	path       string
	body       string
	titleLower string
	haystack   string
}

// SearchIndex is an in-memory index of markdown content.
type SearchIndex struct {
	docs []searchDoc
}

// BuildSearchIndex walks markdown files under contentDir and builds a search index.
func BuildSearchIndex(contentDir string) (*SearchIndex, error) {
	idx := &SearchIndex{}
	cleanDir := filepath.Clean(contentDir)

	err := filepath.WalkDir(cleanDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		rel, relErr := filepath.Rel(cleanDir, path)
		if relErr != nil {
			return relErr
		}
		slug := strings.TrimSuffix(filepath.ToSlash(rel), filepath.Ext(rel))
		route := buildRoutePath(slug)
		if strings.HasSuffix(route, "/index") {
			route = strings.TrimSuffix(route, "/index")
			if route == "" {
				route = "/"
			}
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		meta, body := parseFrontmatter(string(data))

		if override := routeFromFrontmatter(meta); override != "" {
			route = override
		}

		title := strings.TrimSpace(meta["title"])
		if heading, rest, ok := stripLeadingH1(body); ok {
			body = rest
			if title == "" {
				title = heading
			}
		}
		if title == "" {
			title = pageTitleFromSlug(slug)
		}

		bodyText := normalizeText(body)
		idx.docs = append(idx.docs, searchDoc{
			title:      title,
			path:       route,
			body:       bodyText,
			titleLower: strings.ToLower(title),
			haystack:   strings.ToLower(title + " " + bodyText),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return idx, nil
}

func normalizeText(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isWholeWord reports whether term appears in haystack delimited by non-alphanumeric
// boundaries (so "go" matches "go" but not "golang").
func isWholeWord(haystack, term string) bool {
	from := 0
	for {
		i := strings.Index(haystack[from:], term)
		if i == -1 {
			return false
		}
		i += from
		beforeOK := i == 0 || !isWordByte(haystack[i-1])
		end := i + len(term)
		afterOK := end >= len(haystack) || !isWordByte(haystack[end])
		if beforeOK && afterOK {
			return true
		}
		from = i + 1
	}
}

func isWordByte(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
}

func splitTerms(q string) []string {
	q = strings.TrimSpace(strings.ToLower(q))
	if q == "" {
		return nil
	}
	return strings.Fields(q)
}

func makeSnippet(body string, terms []string) string {
	if body == "" {
		return ""
	}

	bodyLower := strings.ToLower(body)
	at := -1
	for _, term := range terms {
		if term == "" {
			continue
		}
		if i := strings.Index(bodyLower, term); i != -1 {
			at = i
			break
		}
	}
	if at == -1 {
		if len(body) <= 180 {
			return body
		}
		return strings.TrimSpace(body[:180]) + "..."
	}

	start := at - 72
	if start < 0 {
		start = 0
	}
	end := start + 180
	if end > len(body) {
		end = len(body)
	}

	prefix := ""
	suffix := ""
	if start > 0 {
		prefix = "..."
	}
	if end < len(body) {
		suffix = "..."
	}

	return prefix + strings.TrimSpace(body[start:end]) + suffix
}

// Query finds markdown pages matching q, ordered by a simple relevance score.
func (idx *SearchIndex) Query(q string, limit int) []SearchResult {
	if idx == nil {
		return nil
	}
	terms := splitTerms(q)
	if len(terms) == 0 {
		return nil
	}
	if limit <= 0 {
		limit = 8
	}

	type scored struct {
		result SearchResult
		score  int
	}

	matches := make([]scored, 0, len(idx.docs))
	queryLower := strings.ToLower(strings.TrimSpace(q))
	for _, doc := range idx.docs {
		score := 0
		allTermsPresent := true
		for _, term := range terms {
			if !strings.Contains(doc.haystack, term) {
				allTermsPresent = false
				break
			}
			// Base credit for the term, plus a capped term-frequency boost so a
			// page that discusses the term repeatedly ranks above an incidental
			// mention.
			score += 2
			if freq := strings.Count(doc.body, term); freq > 0 {
				score += minInt(freq, 5)
			}
			// A term in the title is a much stronger relevance signal than body.
			if strings.Contains(doc.titleLower, term) {
				score += 3
				if isWholeWord(doc.titleLower, term) {
					score += 2
				}
			}
		}
		if !allTermsPresent {
			continue
		}

		// Exact phrase matches rank highest: the full query appearing verbatim in
		// the title, then in the body.
		if strings.Contains(doc.titleLower, queryLower) {
			score += 5
		} else if len(terms) > 1 && strings.Contains(doc.body, queryLower) {
			score += 3
		}

		matches = append(matches, scored{
			result: SearchResult{
				Title:   doc.title,
				Path:    doc.path,
				Snippet: makeSnippet(doc.body, terms),
			},
			score: score,
		})
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		if matches[i].result.Title != matches[j].result.Title {
			return matches[i].result.Title < matches[j].result.Title
		}
		return matches[i].result.Path < matches[j].result.Path
	})

	if limit > len(matches) {
		limit = len(matches)
	}
	out := make([]SearchResult, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, matches[i].result)
	}
	return out
}
