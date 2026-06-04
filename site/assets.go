package site

import (
	"embed"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//go:embed public/*
var embeddedPublicFS embed.FS

// publicFS returns the filesystem serving static public assets: an on-disk
// directory when configured, otherwise the embedded public/ tree.
func (a *App) publicFS() (fs.FS, error) {
	if dir := strings.TrimSpace(a.PublicDir); dir != "" {
		return os.DirFS(filepath.Clean(dir)), nil
	}

	publicFS, err := fs.Sub(embeddedPublicFS, "public")
	if err != nil {
		return nil, err
	}

	return publicFS, nil
}

func staticFileExists(publicFS fs.FS, requestPath string) (bool, error) {
	relReq := strings.TrimPrefix(requestPath, "/")
	relReq = path.Clean(relReq)

	if relReq == "." || relReq == "" || strings.HasPrefix(relReq, "../") || relReq == ".." {
		return false, nil
	}

	info, statErr := fs.Stat(publicFS, relReq)
	if statErr != nil {
		if errors.Is(statErr, fs.ErrNotExist) {
			return false, nil
		}
		return false, statErr
	}

	return !info.IsDir(), nil
}
