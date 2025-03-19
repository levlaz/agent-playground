// Basic workspace module for Daggerverse QA Agent
package main

import (
	"context"
	"dagger/workspace/internal/dagger"
)

type Workspace struct {
	// Workspace container state
	// +internal-use-only
	Container *dagger.Container
}

func New() Workspace {
	return Workspace{
		Container: dag.Container().
			From("alpine:latest").
			WithDirectory("/qa", dag.Directory()).
			WithWorkdir("/qa").
			WithExec([]string{"apk", "add", "curl", "docker"}).
			WithExec([]string{"sh", "-c", "curl -fsSL https://dl.dagger.io/dagger/install.sh | BIN_DIR=/usr/local/bin sh"}).
			WithExec([]string{"dagger", "init"}, dagger.ContainerWithExecOpts{ExperimentalPrivilegedNesting: true}),
	}
}

// Read a file
func (w *Workspace) Read(ctx context.Context, path string) (string, error) {
	return w.Container.File(path).Contents(ctx)
}

// Write a file
func (w *Workspace) Write(path, content string) *Workspace {
	w.Container = w.Container.WithNewFile(path, content)
	return w
}

// Install a module
func (m *Workspace) Install(ctx context.Context, module string) *Workspace {
	m.Container = m.Container.
		WithExec([]string{"dagger", "install", module}, dagger.ContainerWithExecOpts{ExperimentalPrivilegedNesting: true})
	return m
}

// Build a module
func (m *Workspace) Build(ctx context.Context, module string) *Workspace {
	m.Container = m.Container.
		WithExec([]string{"dagger", "-m", module, "functions"}, dagger.ContainerWithExecOpts{ExperimentalPrivilegedNesting: true})
	return m
}
