package gomark

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSidebarRootIncludesRoutableAndToggleOnlyFolders(t *testing.T) {
	contentDir := t.TempDir()
	writeMarkdown(t, contentDir, "index.md", "---\ntitle: Home\n---\n")
	writeMarkdown(t, contentDir, "routable/index.md", "---\ntitle: Routable\n---\n")
	writeMarkdown(t, contentDir, "routable/page.md", "# Page\n")
	writeMarkdown(t, contentDir, "toggle/child.md", "# Child\n")

	idx, err := BuildContentIndex(contentDir)
	if err != nil {
		t.Fatalf("build content index: %v", err)
	}

	title, nodes := idx.Sidebar("/", 3)

	if title != "Home" {
		t.Fatalf("expected root sidebar title from home index, got %q", title)
	}

	toggle, ok := findNavNode(nodes, "Toggle")
	if !ok {
		t.Fatalf("expected Toggle folder in root sidebar")
	}
	if toggle.Path != "" {
		t.Fatalf("expected Toggle to be toggle-only (no route), got %q", toggle.Path)
	}
	if len(toggle.Children) == 0 {
		t.Fatalf("expected Toggle to include children for accordion toggling")
	}

	routable, ok := findNavNode(nodes, "Routable")
	if !ok {
		t.Fatalf("expected Routable folder in root sidebar")
	}
	if routable.Path == "" {
		t.Fatalf("expected Routable to be routable via index.md")
	}
}

func TestSidebarKeepsOffPathChildrenForAccordion(t *testing.T) {
	contentDir := t.TempDir()
	writeMarkdown(t, contentDir, "alpha/index.md", "---\ntitle: Alpha\n---\n")
	writeMarkdown(t, contentDir, "alpha/one/index.md", "---\ntitle: One\n---\n")
	writeMarkdown(t, contentDir, "alpha/one/page.md", "# One Page\n")
	writeMarkdown(t, contentDir, "alpha/two/index.md", "---\ntitle: Two\n---\n")
	writeMarkdown(t, contentDir, "alpha/two/page.md", "# Two Page\n")

	idx, err := BuildContentIndex(contentDir)
	if err != nil {
		t.Fatalf("build content index: %v", err)
	}

	_, nodes := idx.Sidebar("/alpha/one", 3)

	one, ok := findNavNode(nodes, "One")
	if !ok {
		t.Fatalf("expected active One folder")
	}
	if !one.Open {
		t.Fatalf("expected active ancestor folder to be open")
	}

	two, ok := findNavNode(nodes, "Two")
	if !ok {
		t.Fatalf("expected off-path Two folder")
	}
	if two.Open {
		t.Fatalf("expected off-path folder to be closed by default")
	}
	if len(two.Children) == 0 {
		t.Fatalf("expected off-path folder children to be present for accordion rendering")
	}
}

func TestSidebarStaysRootAnchoredOnNestedRoute(t *testing.T) {
	contentDir := t.TempDir()
	writeMarkdown(t, contentDir, "index.md", "---\ntitle: Home\n---\n")
	writeMarkdown(t, contentDir, "patterns/index.md", "---\ntitle: Patterns\n---\n")
	writeMarkdown(t, contentDir, "patterns/creational/index.md", "---\ntitle: Creational\n---\n")
	writeMarkdown(t, contentDir, "philosophy/index.md", "---\ntitle: Philosophy\n---\n")

	idx, err := BuildContentIndex(contentDir)
	if err != nil {
		t.Fatalf("build content index: %v", err)
	}

	title, nodes := idx.Sidebar("/patterns", 3)

	if title != "Home" {
		t.Fatalf("expected root-anchored sidebar title Home, got %q", title)
	}

	patterns, ok := findNavNode(nodes, "Patterns")
	if !ok {
		t.Fatalf("expected Patterns folder in root sidebar")
	}
	if !patterns.Active {
		t.Fatalf("expected Patterns folder to be active on /patterns")
	}
	if !patterns.Open {
		t.Fatalf("expected Patterns folder to be open on /patterns")
	}

	if _, ok := findNavNode(nodes, "Philosophy"); !ok {
		t.Fatalf("expected sibling root folder Philosophy to remain visible")
	}
}

func TestSidebarIncludesRootIndexAsLink(t *testing.T) {
	contentDir := t.TempDir()
	writeMarkdown(t, contentDir, "index.md", "---\ntitle: Home\n---\n")
	writeMarkdown(t, contentDir, "patterns/index.md", "---\ntitle: Patterns\n---\n")

	idx, err := BuildContentIndex(contentDir)
	if err != nil {
		t.Fatalf("build content index: %v", err)
	}

	_, nodes := idx.Sidebar("/patterns", 3)

	home, ok := findNavNode(nodes, "Home")
	if !ok {
		t.Fatalf("expected root index page Home in sidebar nav tree")
	}
	if home.Path != "/" {
		t.Fatalf("expected Home nav link path '/', got %q", home.Path)
	}
}

func TestSidebarUsesNavTitleWhenProvided(t *testing.T) {
	contentDir := t.TempDir()
	writeMarkdown(t, contentDir, "index.md", "---\ntitle: GoMark\nnav_title: Home\n---\n")

	idx, err := BuildContentIndex(contentDir)
	if err != nil {
		t.Fatalf("build content index: %v", err)
	}

	title, nodes := idx.Sidebar("/", 3)

	if title != "GoMark" {
		t.Fatalf("expected page title to remain GoMark, got %q", title)
	}

	home, ok := findNavNode(nodes, "Home")
	if !ok {
		t.Fatalf("expected nav label Home from nav_title")
	}
	if home.Path != "/" {
		t.Fatalf("expected Home nav link path '/', got %q", home.Path)
	}

	if _, ok := findNavNode(nodes, "GoMark"); ok {
		t.Fatalf("did not expect page title to appear as nav label when nav_title is set")
	}
}

func TestChildPagesSortByOrder(t *testing.T) {
	contentDir := t.TempDir()
	writeMarkdown(t, contentDir, "second.md", "---\ntitle: Second\norder: 2\n---\n")
	writeMarkdown(t, contentDir, "first.md", "---\ntitle: First\norder: 1\n---\n")
	writeMarkdown(t, contentDir, "last.md", "---\ntitle: Last\n---\n")

	idx, err := BuildContentIndex(contentDir)
	if err != nil {
		t.Fatalf("build content index: %v", err)
	}

	_, nodes := idx.Sidebar("/", 3)

	got := navTitlesInOrder(nodes)
	want := []string{"First", "Second", "Last"}
	if !equalStrings(got, want) {
		t.Fatalf("expected sidebar order %v, got %v", want, got)
	}
}

func TestUnorderedPagesSortLastThenByTitle(t *testing.T) {
	contentDir := t.TempDir()
	writeMarkdown(t, contentDir, "ordered.md", "---\ntitle: Ordered\norder: 1\n---\n")
	writeMarkdown(t, contentDir, "banana.md", "---\ntitle: Banana\n---\n")
	writeMarkdown(t, contentDir, "apple.md", "---\ntitle: Apple\n---\n")

	idx, err := BuildContentIndex(contentDir)
	if err != nil {
		t.Fatalf("build content index: %v", err)
	}

	_, nodes := idx.Sidebar("/", 3)

	got := navTitlesInOrder(nodes)
	want := []string{"Ordered", "Apple", "Banana"}
	if !equalStrings(got, want) {
		t.Fatalf("expected unordered pages last, alphabetized: %v, got %v", want, got)
	}
}

func TestChildDirsSortByIndexOrder(t *testing.T) {
	contentDir := t.TempDir()
	writeMarkdown(t, contentDir, "zebra/index.md", "---\ntitle: Zebra\norder: 1\n---\n")
	writeMarkdown(t, contentDir, "alpha/index.md", "---\ntitle: Alpha\norder: 2\n---\n")

	idx, err := BuildContentIndex(contentDir)
	if err != nil {
		t.Fatalf("build content index: %v", err)
	}

	_, nodes := idx.Sidebar("/", 3)

	got := navTitlesInOrder(nodes)
	want := []string{"Zebra", "Alpha"}
	if !equalStrings(got, want) {
		t.Fatalf("expected folders ordered by index.md order %v, got %v", want, got)
	}
}

func TestDirOrderInferredFromChildrenWhenNoIndex(t *testing.T) {
	contentDir := t.TempDir()
	// Folder without index.md whose only page has a low order.
	writeMarkdown(t, contentDir, "noindex/page.md", "---\ntitle: Page\norder: 1\n---\n")
	// Indexed folder with a higher order.
	writeMarkdown(t, contentDir, "indexed/index.md", "---\ntitle: Indexed\norder: 5\n---\n")

	idx, err := BuildContentIndex(contentDir)
	if err != nil {
		t.Fatalf("build content index: %v", err)
	}

	_, nodes := idx.Sidebar("/", 3)

	got := navTitlesInOrder(nodes)
	want := []string{"Noindex", "Indexed"}
	if !equalStrings(got, want) {
		t.Fatalf("expected folder order inferred from children %v, got %v", want, got)
	}
}

func TestPagesAndDirsInterleaveByOrder(t *testing.T) {
	contentDir := t.TempDir()
	// A low-order root page should sort above higher-order folders.
	writeMarkdown(t, contentDir, "playground.md", "---\ntitle: Playground\norder: 1\n---\n")
	writeMarkdown(t, contentDir, "guides/index.md", "---\ntitle: Guides\norder: 2\n---\n")
	writeMarkdown(t, contentDir, "guides/page.md", "# Guide Page\n")
	writeMarkdown(t, contentDir, "reference/index.md", "---\ntitle: Reference\norder: 3\n---\n")
	writeMarkdown(t, contentDir, "reference/page.md", "# Reference Page\n")

	idx, err := BuildContentIndex(contentDir)
	if err != nil {
		t.Fatalf("build content index: %v", err)
	}

	_, nodes := idx.Sidebar("/", 3)
	got := navTitlesInOrder(nodes)
	want := []string{"Playground", "Guides", "Reference"}
	if !equalStrings(got, want) {
		t.Fatalf("expected pages and dirs interleaved by order %v, got %v", want, got)
	}

	topGot := navLinkTitlesInOrder(idx.TopNav())
	if !equalStrings(topGot, want) {
		t.Fatalf("expected top nav interleaved by order %v, got %v", want, topGot)
	}
}

func writeMarkdown(t *testing.T, root, relPath, content string) {
	t.Helper()
	fullPath := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", relPath, err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", relPath, err)
	}
}

func navTitlesInOrder(nodes []NavNode) []string {
	titles := make([]string, 0, len(nodes))
	for _, node := range nodes {
		titles = append(titles, node.Title)
	}
	return titles
}

func navLinkTitlesInOrder(links []NavLink) []string {
	titles := make([]string, 0, len(links))
	for _, link := range links {
		titles = append(titles, link.Title)
	}
	return titles
}

func TestSiblingsOrderedByFrontmatterOrder(t *testing.T) {
	contentDir := t.TempDir()
	writeMarkdown(t, contentDir, "guide/index.md", "---\ntitle: Guide\n---\n")
	writeMarkdown(t, contentDir, "guide/second.md", "---\ntitle: Second\norder: 2\n---\n")
	writeMarkdown(t, contentDir, "guide/first.md", "---\ntitle: First\norder: 1\n---\n")
	writeMarkdown(t, contentDir, "guide/third.md", "---\ntitle: Third\norder: 3\n---\n")

	idx, err := BuildContentIndex(contentDir)
	if err != nil {
		t.Fatalf("build content index: %v", err)
	}

	prev, next := idx.Siblings("/guide/second")
	if prev == nil || prev.Title != "First" {
		t.Fatalf("expected previous sibling First, got %v", prev)
	}
	if next == nil || next.Title != "Third" {
		t.Fatalf("expected next sibling Third, got %v", next)
	}

	if prev, _ := idx.Siblings("/guide/first"); prev != nil {
		t.Fatalf("expected no previous sibling for the first page, got %v", prev)
	}
	if _, next := idx.Siblings("/guide/third"); next != nil {
		t.Fatalf("expected no next sibling for the last page, got %v", next)
	}
}

func TestSiblingsFallBackToAlphabeticalOrder(t *testing.T) {
	contentDir := t.TempDir()
	writeMarkdown(t, contentDir, "apple.md", "---\ntitle: Apple\n---\n")
	writeMarkdown(t, contentDir, "banana.md", "---\ntitle: Banana\n---\n")
	writeMarkdown(t, contentDir, "cherry.md", "---\ntitle: Cherry\n---\n")

	idx, err := BuildContentIndex(contentDir)
	if err != nil {
		t.Fatalf("build content index: %v", err)
	}

	prev, next := idx.Siblings("/banana")
	if prev == nil || prev.Title != "Apple" {
		t.Fatalf("expected previous sibling Apple, got %v", prev)
	}
	if next == nil || next.Title != "Cherry" {
		t.Fatalf("expected next sibling Cherry, got %v", next)
	}
}

func TestSiblingsReturnsNilForUnknownOrIndexRoutes(t *testing.T) {
	contentDir := t.TempDir()
	writeMarkdown(t, contentDir, "index.md", "---\ntitle: Home\n---\n")
	writeMarkdown(t, contentDir, "guide/index.md", "---\ntitle: Guide\n---\n")
	writeMarkdown(t, contentDir, "guide/page.md", "---\ntitle: Page\n---\n")

	idx, err := BuildContentIndex(contentDir)
	if err != nil {
		t.Fatalf("build content index: %v", err)
	}

	if prev, next := idx.Siblings("/does-not-exist"); prev != nil || next != nil {
		t.Fatalf("expected nil siblings for unknown route, got prev=%v next=%v", prev, next)
	}
	if prev, next := idx.Siblings("/guide"); prev != nil || next != nil {
		t.Fatalf("expected nil siblings for an index/folder route, got prev=%v next=%v", prev, next)
	}
}

func TestNavNodeIconFromFrontmatter(t *testing.T) {
	contentDir := t.TempDir()
	writeMarkdown(t, contentDir, "index.md", "---\ntitle: Home\n---\n")
	writeMarkdown(t, contentDir, "rocket.md", "---\ntitle: Rocket\nicon: rocket\n---\n")
	writeMarkdown(t, contentDir, "guide/index.md", "---\ntitle: Guide\nicon: book\n---\n")

	idx, err := BuildContentIndex(contentDir)
	if err != nil {
		t.Fatalf("build content index: %v", err)
	}

	_, nodes := idx.Sidebar("/", 3)

	page, ok := findNavNode(nodes, "Rocket")
	if !ok {
		t.Fatalf("expected Rocket page in sidebar")
	}
	if page.Icon != "rocket" {
		t.Fatalf("expected page icon %q, got %q", "rocket", page.Icon)
	}

	folder, ok := findNavNode(nodes, "Guide")
	if !ok {
		t.Fatalf("expected Guide folder in sidebar")
	}
	if folder.Icon != "book" {
		t.Fatalf("expected folder icon %q from its index.md, got %q", "book", folder.Icon)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func findNavNode(nodes []NavNode, title string) (NavNode, bool) {
	for _, node := range nodes {
		if node.Title == title {
			return node, true
		}
		if child, ok := findNavNode(node.Children, title); ok {
			return child, true
		}
	}
	return NavNode{}, false
}
