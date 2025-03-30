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

type DaggerverseQa struct{}

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
		WithExec([]string{"sh", "-c", "cat modules.json | jq -r '.[].path' | sort | uniq  | shuf | head -n 3"}).
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

	return output, nil
}

// Perform Single QA Run
func (m *DaggerverseQa) Run(ctx context.Context, module string) (*dagger.Directory, error) {
	before := dag.Workspace()

	after := dag.LLM().
		WithWorkspace(before).
		WithPromptVar("modules", module).
		WithPromptFile(dag.CurrentModule().Source().File("qa.prompt")).
		Workspace()

	return after.Container().Directory("/qa"), nil
}
