//go:build !darwin

package main

import "context"

// beforeClose is only needed on macOS; elsewhere Wails quits on close by itself.
var beforeClose func(context.Context) bool
