package gomark

import (
	"io/fs"
	"strings"
	"testing"
)

func TestAppPublicFSDefaultsToEmbeddedAssets(t *testing.T) {
	a := App{}
	publicFS, err := a.publicFS()
	if err != nil {
		t.Fatalf("public fs: %v", err)
	}

	entries, err := fs.ReadDir(publicFS, ".")
	if err != nil {
		t.Fatalf("read embedded public dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected embedded public assets")
	}

	foundFavicon := false
	for _, entry := range entries {
		if strings.EqualFold(entry.Name(), "favicon.ico") {
			foundFavicon = true
			break
		}
	}
	if !foundFavicon {
		t.Fatal("expected embedded favicon.ico")
	}
}

func TestStaticFileExistsUsesFSAndBlocksTraversal(t *testing.T) {
	a := App{}
	publicFS, err := a.publicFS()
	if err != nil {
		t.Fatalf("public fs: %v", err)
	}

	exists, err := staticFileExists(publicFS, "/favicon.ico")
	if err != nil {
		t.Fatalf("static file exists: %v", err)
	}
	if !exists {
		t.Fatal("expected favicon.ico to exist")
	}

	exists, err = staticFileExists(publicFS, "/../runtime.go")
	if err != nil {
		t.Fatalf("static file traversal check: %v", err)
	}
	if exists {
		t.Fatal("did not expect path traversal to resolve")
	}
}
