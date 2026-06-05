package gomark

import (
	"embed"
	"errors"
	"io/fs"
	"path"
	"strings"
)

//go:embed public/*
var embeddedPublicFS embed.FS

// publicFS returns the embedded public/ tree that serves static assets
// (favicons, og images, vendored JS/CSS, the runner module).
func (a *App) publicFS() (fs.FS, error) {
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
