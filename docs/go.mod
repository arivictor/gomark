// The gomark.dev documentation site, built as a downstream consumer of the
// gomark module itself. It depends on github.com/arivictor/gomark and renders
// this directory with the gomark CLI — exactly as an external user would —
// while the replace below points the dependency at this repo's working tree so
// the site always builds against the in-tree version of gomark.
//
// Build it (no Go code required):
//
//	cd docs && go tool gomark build
module github.com/arivictor/gomark/docs

go 1.24.4

replace github.com/arivictor/gomark => ../

tool github.com/arivictor/gomark/cmd/gomark

require (
	github.com/arivictor/gomark v0.1.12 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
