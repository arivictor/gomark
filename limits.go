package gomark

import "time"

const (
	MaxSourceBytes = 64 << 10
	MaxOutputBytes = 64 << 10
	RunTimeout     = 2 * time.Second
)
