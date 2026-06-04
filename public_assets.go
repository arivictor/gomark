package gomark

import "embed"

//go:embed public/*
var embeddedPublicFS embed.FS
