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

// Attempt to install a module and get the result
func (m *Workspace) Install(ctx context.Context, module string, version string) (string, error) {
	return m.Container.
		WithExec([]string{"dagger", "install", module}, dagger.ContainerWithExecOpts{ExperimentalPrivilegedNesting: true}).Stdout(ctx)
}

// Build a module and list its functions
func (m *Workspace) Build(ctx context.Context, module string, version string) (string, error) {
	return m.Container.
		WithExec([]string{"dagger", "-m", module, "functions"}, dagger.ContainerWithExecOpts{ExperimentalPrivilegedNesting: true}).Stdout(ctx)
}

// Get Dagger Version
func (m *Workspace) Version(ctx context.Context) (string, error) {
	return m.Container.WithExec([]string{"dagger", "version"}).Stdout(ctx)
}
