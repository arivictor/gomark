package gomark

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const unordered = 1 << 30

// NavNode is one entry in the sidebar tree. Folders may carry Children; pages are
// leaves. Active marks the exact current page; Open marks an ancestor of it.
type NavNode struct {
	Title    string
	Path     string // link target; empty for a folder without an index.md
	Icon     string // optional Lucide icon name from frontmatter
	NodeID   string // stable id for folder accordion controls
	Folder   bool   // a directory (styled as a header) vs. a leaf page
	Active   bool
	Open     bool
	Children []NavNode
}

// NavLink is a single entry in the header's top-level nav.
type NavLink struct {
	Title  string
	Path   string
	Active bool
}

// contentPage is one markdown page in the in-memory content index.
type contentPage struct {
	Title    string
	NavTitle string
	Icon     string   // optional Lucide icon name from frontmatter "icon"/"nav_icon"
	Route    string   // clean route; index pages map to their folder path ("/", "/go")
	Parts    []string // slug segments, e.g. ["go","patterns","creational","singleton"]
	IsIndex  bool
	Order    int // frontmatter "order"; unordered pages sort last
}

// dirEntry is a sub-directory discovered while grouping, with metadata pulled
// from its index.md when present.
type dirEntry struct {
	Name  string
	Title string
	Icon  string
	Route string // index route if the dir has an index.md, else empty
	Order int
}

// ContentIndex holds every page, built once at startup, and renders the
// position-relative sidebar for a given route.
type ContentIndex struct {
	pages []contentPage
}

// BuildContentIndex walks the content dir and records each page's title (from
// frontmatter) and position, so the sidebar can be built without re-reading files.
func BuildContentIndex(dir string) (*ContentIndex, error) {
	idx := &ContentIndex{}
	cleanDir := filepath.Clean(dir)

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
		parts := strings.Split(slug, "/")

		isIndex := len(parts) > 0 && parts[len(parts)-1] == "index"
		if isIndex {
			parts = parts[:len(parts)-1]
			route = cleanRoute(route)
		}

		data, _ := os.ReadFile(path)
		meta, _ := parseFrontmatter(string(data))

		if override := routeFromFrontmatter(meta); override != "" {
			route = override
		}

		title := strings.TrimSpace(meta["title"])
		if title == "" {
			title = pageTitleFromSlug(slug)
		}
		navTitle := strings.TrimSpace(firstNonEmpty(meta["nav_title"], meta["navtitle"], meta["nav"]))
		if navTitle == "" {
			navTitle = title
		}
		icon := strings.TrimSpace(firstNonEmpty(meta["icon"], meta["nav_icon"]))
		order := unordered
		if o, convErr := strconv.Atoi(strings.TrimSpace(meta["order"])); convErr == nil {
			order = o
		}

		idx.pages = append(idx.pages, contentPage{
			Title:    title,
			NavTitle: navTitle,
			Icon:     icon,
			Route:    route,
			Parts:    parts,
			IsIndex:  isIndex,
			Order:    order,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return idx, nil
}

// TopNav returns the content root's top-level entries for the header: each
// top-level folder that has an index.md plus any root-level pages. The root
// index.md itself is omitted (it's the brand/home link).
func (idx *ContentIndex) TopNav() []NavLink {
	type ordered struct {
		order int
		title string
		link  NavLink
	}
	var entries []ordered
	for _, dir := range idx.childDirs(nil) {
		if dir.Route == "" {
			continue
		}
		entries = append(entries, ordered{
			order: dir.Order,
			title: dir.Title,
			link:  NavLink{Title: dir.Title, Path: dir.Route},
		})
	}
	for _, p := range idx.childPages(nil) {
		entries = append(entries, ordered{
			order: p.Order,
			title: p.NavTitle,
			link:  NavLink{Title: p.NavTitle, Path: p.Route},
		})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].order != entries[j].order {
			return entries[i].order < entries[j].order
		}
		return entries[i].title < entries[j].title
	})
	links := make([]NavLink, 0, len(entries))
	for _, e := range entries {
		links = append(links, e.link)
	}
	return links
}

// Sidebar returns the heading and node tree anchored at content root for every
// route, so navigating into a folder keeps the same overall accordion tree.
func (idx *ContentIndex) Sidebar(currentRoute string, depth int) (string, []NavNode) {
	cur := cleanRoute(currentRoute)
	current, _ := idx.byRoute(cur)
	rootTitle := idx.dirTitle(nil, "Home")
	if strings.TrimSpace(rootTitle) == "" {
		rootTitle = "Home"
	}
	return rootTitle, idx.buildNodes(nil, current.Parts, cur, depth)
}

// Siblings returns the "previous" and "next" entries for route, drawn from its
// sibling pages (same parent folder), in the same order the sidebar lists them
// (frontmatter "order", falling back to alphabetical title). Either may be nil
// at the ends of the list, or if the route isn't a recognized page.
func (idx *ContentIndex) Siblings(route string) (prev, next *NavLink) {
	cur, ok := idx.byRoute(cleanRoute(route))
	if !ok || len(cur.Parts) == 0 {
		return nil, nil
	}

	prefix := cur.Parts[:len(cur.Parts)-1]
	siblings := idx.childPages(prefix)
	sort.SliceStable(siblings, func(i, j int) bool {
		if siblings[i].Order != siblings[j].Order {
			return siblings[i].Order < siblings[j].Order
		}
		return siblings[i].NavTitle < siblings[j].NavTitle
	})

	for i, p := range siblings {
		if p.Route != cur.Route {
			continue
		}
		if i > 0 {
			s := siblings[i-1]
			prev = &NavLink{Title: s.NavTitle, Path: s.Route}
		}
		if i+1 < len(siblings) {
			s := siblings[i+1]
			next = &NavLink{Title: s.NavTitle, Path: s.Route}
		}
		return prev, next
	}

	return nil, nil
}

// buildNodes returns the nav nodes for the children of prefix: sub-folders
// (expanded only when on the current path) and pages, interleaved by order then
// title so a page and a folder compete for position on equal footing.
func (idx *ContentIndex) buildNodes(prefix, curParts []string, cur string, depth int) []NavNode {
	var nodes []NavNode
	if len(prefix) == 0 {
		if root, ok := idx.indexFor(nil); ok {
			nodes = append(nodes, NavNode{Title: root.NavTitle, Path: root.Route, Icon: root.Icon, Active: root.Route == cur})
		}
	}

	type ordered struct {
		order int
		title string
		node  NavNode
	}
	var entries []ordered

	for _, dir := range idx.childDirs(prefix) {
		childPrefix := append(append([]string{}, prefix...), dir.Name)
		onPath := hasPrefix(curParts, childPrefix)
		children := idx.buildNodes(childPrefix, curParts, cur, depth)

		entries = append(entries, ordered{
			order: dir.Order,
			title: dir.Title,
			node: NavNode{
				Title:    dir.Title,
				Path:     dir.Route,
				Icon:     dir.Icon,
				NodeID:   navNodeID(childPrefix),
				Folder:   true,
				Active:   dir.Route == cur,
				Open:     onPath,
				Children: children,
			},
		})
	}
	for _, p := range idx.childPages(prefix) {
		entries = append(entries, ordered{
			order: p.Order,
			title: p.NavTitle,
			node:  NavNode{Title: p.NavTitle, Path: p.Route, Icon: p.Icon, Active: p.Route == cur},
		})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].order != entries[j].order {
			return entries[i].order < entries[j].order
		}
		return entries[i].title < entries[j].title
	})
	for _, e := range entries {
		nodes = append(nodes, e.node)
	}
	return nodes
}

// childPages returns the non-index pages directly inside prefix, sorted by order
// then title.
func (idx *ContentIndex) childPages(prefix []string) []contentPage {
	var out []contentPage
	for _, p := range idx.pages {
		if p.IsIndex || len(p.Parts) != len(prefix)+1 || !hasPrefix(p.Parts, prefix) {
			continue
		}
		out = append(out, p)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].Title < out[j].Title
	})
	return out
}

// childDirs returns the sub-directories one level below prefix. A dir's title,
// route, and order come from its index.md when present; otherwise the order is
// inferred from the smallest order among its pages so unindexed folders still
// sort sensibly.
func (idx *ContentIndex) childDirs(prefix []string) []dirEntry {
	seen := map[string]*dirEntry{}
	isDir := map[string]bool{}
	var order []string

	for _, p := range idx.pages {
		if len(p.Parts) <= len(prefix) || !hasPrefix(p.Parts, prefix) {
			continue
		}
		name := p.Parts[len(prefix)]
		entry, ok := seen[name]
		if !ok {
			entry = &dirEntry{Name: name, Title: humanizeSlug(name), Order: unordered}
			seen[name] = entry
			order = append(order, name)
		}
		if p.Order < entry.Order {
			entry.Order = p.Order
		}
		// `name` is a directory only if something lives below it or it has its own
		// index.md — a leaf page directly under prefix is not a directory.
		if len(p.Parts) > len(prefix)+1 {
			isDir[name] = true
		}
		if p.IsIndex && len(p.Parts) == len(prefix)+1 {
			isDir[name] = true
			entry.Title = p.NavTitle
			entry.Icon = p.Icon
			entry.Route = p.Route
		}
	}

	out := make([]dirEntry, 0, len(order))
	for _, name := range order {
		if !isDir[name] {
			continue
		}
		out = append(out, *seen[name])
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].Title < out[j].Title
	})
	return out
}

func (idx *ContentIndex) byRoute(route string) (contentPage, bool) {
	for _, p := range idx.pages {
		if p.Route == route {
			return p, true
		}
	}
	return contentPage{}, false
}

func (idx *ContentIndex) indexFor(prefix []string) (contentPage, bool) {
	for _, p := range idx.pages {
		if p.IsIndex && len(p.Parts) == len(prefix) && hasPrefix(p.Parts, prefix) {
			return p, true
		}
	}
	return contentPage{}, false
}

func (idx *ContentIndex) dirTitle(prefix []string, fallback string) string {
	if p, ok := idx.indexFor(prefix); ok && strings.TrimSpace(p.Title) != "" {
		return p.Title
	}
	return fallback
}

func hasPrefix(parts, prefix []string) bool {
	if len(parts) < len(prefix) {
		return false
	}
	for i := range prefix {
		if parts[i] != prefix[i] {
			return false
		}
	}
	return true
}

// cleanRoute maps an index route to its folder path: "/go/index" -> "/go" and
// "/index" -> "/".
func cleanRoute(route string) string {
	r := strings.TrimSuffix(route, "/index")
	if r == "" {
		return "/"
	}
	return r
}

func humanizeSlug(slug string) string {
	return pageTitleFromSlug(slug)
}

func navNodeID(parts []string) string {
	if len(parts) == 0 {
		return "nav-root"
	}
	return "nav-" + slugify(strings.Join(parts, "-"))
}
