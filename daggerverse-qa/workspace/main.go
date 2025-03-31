// Basic workspace module for Daggerverse QA Agent
package main

import (
	"context"
	"dagger/workspace/internal/dagger"
	"fmt"
)

type Workspace struct {
	// Workspace container state
	// +internal-use-only
	Container *dagger.Container
	// Firecrawl token
	FirecrawlToken *dagger.Secret
	// Surge Login
	Login string
	// Surge Token
	SurgeToken *dagger.Secret
	// Surge Domain
	Domain string
}

func New(token *dagger.Secret, login string, surgeToken *dagger.Secret, domain string) Workspace {
	return Workspace{

		Container: dag.Container().
			From("alpine:latest").
			WithExec([]string{"apk", "add", "curl", "docker", "git"}).
			WithExec([]string{"git", "clone", "https://github.com/levlaz/daggerverse-qa-reports", "/qa"}).
			WithWorkdir("/qa").
			WithExec([]string{"sh", "-c", "curl -fsSL https://dl.dagger.io/dagger/install.sh | BIN_DIR=/usr/local/bin sh"}).
			WithExec([]string{"dagger", "init"}, dagger.ContainerWithExecOpts{ExperimentalPrivilegedNesting: true}),
		FirecrawlToken: token,
		Login:          login,
		SurgeToken:     surgeToken,
		Domain:         domain,
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

// Crawl daggerverse page for a module and get vital info
func (m *Workspace) Crawl(ctx context.Context, module string) (string, error) {
	url := fmt.Sprintf("https://daggerverse.dev/mod/%s", module)
	resp, err := dag.FirecrawlDag(m.FirecrawlToken).Scrape(ctx, url)
	if err != nil {
		return "", fmt.Errorf("failed to crawl %s: %v", module, err)
	}
	return resp, nil
}

// Publish a QA reports to surge.sh
func (m *Workspace) Publish(ctx context.Context) (string, error) {
	return dag.Surge(dagger.SurgeOpts{
		Login:   m.Login,
		Token:   m.SurgeToken,
		Domain:  m.Domain,
		Project: m.Container.Directory("/qa"),
	}).Publish().Stdout(ctx)
}

// Get Dagger Version
func (m *Workspace) Version(ctx context.Context) (string, error) {
	return m.Container.WithExec([]string{"dagger", "version"}).Stdout(ctx)
}
