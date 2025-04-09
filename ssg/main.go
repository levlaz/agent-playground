// An agent environment for bulding a static site

package main

import (
	"context"
	"dagger/ssg/internal/dagger"
)

type Ssg struct {
	// Workspace container state
	// +internal-use-only
	Container *dagger.Container
}

// Read a file
func (w *Ssg) Read(ctx context.Context, path string) (string, error) {
	return w.Container.File(path).Contents(ctx)
}

// Write a file
func (w *Ssg) Write(path, content string) *Ssg {
	w.Container = w.Container.WithNewFile(path, content)
	return w
}
