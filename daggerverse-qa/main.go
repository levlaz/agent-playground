// daggerverse-qa is a dagger module that contains functions that perform
// automated QA in the daggerverse. It is an AI Agent powered by Dagger.

package main

import (
	"context"
	"dagger/daggerverse-qa/internal/dagger"
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
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithExec([]string{"apk", "add", "curl", "jq"}).
		WithExec([]string{"sh", "-c", "cat modules.json | jq -r '.[].path' | sort | uniq  | shuf | head -n 1"}).
		Stdout(ctx)
}

// Do QA
func (m *DaggerverseQa) DoQA(
	ctx context.Context,
	// Optional module to test
	// +optional
	modules string,
) *dagger.Container {
	before := dag.Workspace()

	if modules == "" {
		var err error
		modules, err = m.Sample(ctx)
		if err != nil {
			panic(err)
		}
	}

	after := dag.LLM().
		SetWorkspace("workspace", before).
		WithPromptVar("modules", modules).
		WithPromptFile(dag.CurrentModule().Source().File("qa.prompt")).
		GetWorkspace("workspace")

	return after.Container()
}
