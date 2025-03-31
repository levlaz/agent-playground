// daggerverse-qa is a dagger module that contains functions that perform
// automated QA in the daggerverse. It is an AI Agent powered by Dagger.

package main

import (
	"context"
	"dagger/daggerverse-qa/internal/dagger"
	"fmt"
	"strings"
	"time"
)

type DaggerverseQa struct {
	// Surge Login
	Login string
	// Surge Domain
	Domain string
	// Firecrawl token
	FirecrawlToken *dagger.Secret
	// Surge Token
	SurgeToken *dagger.Secret
	// GitHub Token
	GitHubToken *dagger.Secret
}

func New(
	login string,
	domain string,
	firecrawlToken *dagger.Secret,
	surgeToken *dagger.Secret,
	githubToken *dagger.Secret,
) DaggerverseQa {
	return DaggerverseQa{
		Login:          login,
		Domain:         domain,
		FirecrawlToken: firecrawlToken,
		SurgeToken:     surgeToken,
		GitHubToken:    githubToken,
	}
}

// Return list of dagger modules and their latest versions
func (m *DaggerverseQa) Modules(ctx context.Context) *dagger.File {
	return dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "curl", "jq"}).
		WithExec([]string{"sh", "-c", "curl https://daggerverse.dev/api/refs | jq 'group_by(.path) | map({path: .[0].path, latest: (sort_by(.indexed_at) | reverse | .[0])})' > /tmp/modules.json"}).
		File("/tmp/modules.json")
}

// Get a sample of modules from a JSON file
func (m *DaggerverseQa) Sample(ctx context.Context) (string, error) {
	return dag.Container().
		From("alpine:latest").
		WithFile("modules.json", m.Modules(ctx)).
		WithExec([]string{"apk", "add", "curl", "jq"}).
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithExec([]string{"sh", "-c", "cat modules.json | jq -r '.[].path' | sort | uniq  | shuf | head -n 1"}).
		Stdout(ctx)
}

// Iterate over modules and perform QA on each one
func (m *DaggerverseQa) DoQA(
	ctx context.Context,
	// Optional module to test
	// +optional
	modules string,
) (*dagger.Directory, error) {

	if modules == "" {
		var err error
		modules, err = m.Sample(ctx)
		if err != nil {
			return nil, err
		}
	}

	output := dag.Directory()

	for _, module := range strings.Split(modules, " ") {
		report, err := m.Run(ctx, module)
		if err != nil {
			fmt.Errorf("failed to run QA for module %s: %v", module, err)
		}

		output = output.WithDirectory(".", report)
	}

	// Push changes back to GitHub
	dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "git"}).
		WithExec([]string{"git", "config", "--global", "user.name", "Dagger QA agent"}).
		WithExec([]string{"git", "config", "--global", "user.email", "lev@dagger.io"}).
		WithSecretVariable("GH_DQA", m.GitHubToken).
		WithDirectory("/qa", output).
		WithWorkdir("/qa").
		WithExec([]string{"git", "add", "."}).
		WithExec([]string{"git", "commit", "-m", "publish updated QA report"}).
		WithExec([]string{
			"git",
			"push",
			"https://levlaz:$GH_DQA@github.com/levlaz/daggerverse-qa-reports.git",
			"main",
		}).Sync(ctx)

	return output, nil
}

// Perform Single QA Run
func (m *DaggerverseQa) Run(ctx context.Context, module string) (*dagger.Directory, error) {
	before := dag.Workspace(m.FirecrawlToken, m.Login, m.SurgeToken, m.Domain)

	after := dag.LLM().
		WithWorkspace(before).
		WithPromptVar("modules", module).
		WithPromptFile(dag.CurrentModule().Source().File("qa.prompt")).
		Workspace()

	return after.Container().Directory("/qa"), nil
}
