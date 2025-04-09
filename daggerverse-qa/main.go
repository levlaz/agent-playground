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
	FirecrawlToken *dagger.Secret
	GitHubToken    *dagger.Secret
}

func New(
	// Firecrawl token
	// +optional
	firecrawlToken *dagger.Secret,
	// GitHub token
	// +optional
	githubToken *dagger.Secret,
) DaggerverseQa {
	return DaggerverseQa{
		FirecrawlToken: firecrawlToken,
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

	index, indexErr := m.BuildIndex(ctx, output)

	if indexErr != nil {
		return nil, fmt.Errorf("failed to build index: %v", indexErr)
	}

	output = output.WithFile("index.html", index)

	// Push changes back to GitHub
	_, err := m.Push(ctx, output, modules)
	if err != nil {
		return nil, fmt.Errorf("failed to push changes to GitHub: %v", err)
	}

	return output, nil
}

// Push changes back to GitHub
func (m *DaggerverseQa) Push(ctx context.Context, directory *dagger.Directory, modules string) (*dagger.Container, error) {
	return dag.Container().
		From("alpine:latest").
		WithSecretVariable("GITHUB_TOKEN", m.GitHubToken).
		WithExec([]string{"apk", "add", "git"}).
		WithExec([]string{"git", "config", "--global", "user.name", "Dagger QA agent"}).
		WithExec([]string{"git", "config", "--global", "user.email", "lev@dagger.io"}).
		WithDirectory("/qa", directory).
		WithWorkdir("/qa").
		WithExec([]string{"sh", "-c", "git remote set-url origin https://$GITHUB_TOKEN@github.com/levlaz/daggerverse-qa-reports"}).
		WithExec([]string{"git", "add", "."}).
		WithExec([]string{"git", "commit", "-m", fmt.Sprintf("publish updated QA report for %s", modules)}).
		WithExec([]string{
			"git",
			"push",
			"origin",
			"main",
		}).Sync(ctx)
}

// Perform Single QA Run
func (m *DaggerverseQa) Run(ctx context.Context, module string) (*dagger.Directory, error) {
	workspace := dag.Workspace(m.FirecrawlToken)
	environment := dag.Env().
		WithWorkspaceInput("before", workspace, "tools to complete the assignment").
		WithStringInput("module", module, "the module to perform qa on").
		WithWorkspaceOutput("after", "the completed assignment")

	return dag.LLM().
		WithEnv(environment).
		WithPromptFile(dag.CurrentModule().Source().File("qa.prompt")).
		Env().
		Output("after").
		AsWorkspace().
		Container().
		Directory("/qa"), nil
}

// Build an Index Page given a directory full of HTML files
func (m *DaggerverseQa) BuildIndex(ctx context.Context, directory *dagger.Directory) (*dagger.File, error) {
	environment := dag.Env().
		WithDirectoryInput("reports", directory, "the directory to build an index page for").
		WithFileOutput("index", "the index for the directory")

	return dag.LLM().
		WithEnv(environment).
		WithPromptFile(dag.CurrentModule().Source().File("index.prompt")).
		Env().
		Output("index").
		AsFile(), nil
}
