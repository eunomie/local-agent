package main

import (
	"context"
	"dagger/local-agent/internal/dagger"
	"fmt"
)

type LocalAgent struct{}

// Creates a development environment for the project, installs all the needed tools and libraries
func (l *LocalAgent) DevEnvironment(
// Codebase to work on
	source *dagger.Directory,
) *dagger.Container {
	// Create an environment around the source directory:
	env := dag.Env().
		// an alpine based workspace in input, containing:
		// - an alpine based container with the source directory mounted as the input
		// - a set of tools available to the LLM to read files, install packages, etc.
		WithLocalAgentWorkspaceInput(
			"workspace",
			dag.LocalAgent().Workspace(source),
			"An alpine workspace containing the source code directory.").
		// a development environment in output, with all the tools installed
		WithLocalAgentWorkspaceOutput(
			"result",
			"The updated alpine workspace with the necessary development tools and project dependencies installed, based on the analyzed source directory")

	return dag.LLM().
		WithEnv(env).
		WithPrompt("do what you need to do").
		Env().Output("result").AsLocalAgentWorkspace().Container()
}

func (l *LocalAgent) Workspace(source *dagger.Directory) *Workspace {
	return &Workspace{
		Container: dag.Container().
			From("alpine:3").
			WithMountedDirectory("/workspace", source).
			WithWorkdir("/workspace"),
	}
}

// Alpine based workspace
type Workspace struct {
	Container *dagger.Container
}

// Install system packages using apk to the alpine workspace.
//
// Use this to install system packages like `python3`, `git`, etc.
// You cannot install project dependencies with this tool.
func (w *Workspace) AddPackages(
	ctx context.Context,
// List of alpine packages to install
	packages ...string,
) (*Workspace, error) {
	args := append([]string{"apk", "add", "--no-cache"}, packages...)
	return w.WithExec(ctx, args)
}

// Run any command inside the alpine workspace
//
// Use this to install project dependencies, run tests, etc.
func (w *Workspace) WithExec(
	ctx context.Context,
// Command to run
	args []string,
) (*Workspace, error) {
	w.Container = w.Container.WithExec(args, dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeAny})
	exitCode, err := w.Container.ExitCode(ctx)
	if err != nil {
		return w, err
	}
	if exitCode == 0 {
		return w, nil
	}
	stdout, err := w.Container.Stdout(ctx)
	if err != nil {
		return w, err
	}
	stderr, err := w.Container.Stderr(ctx)
	if err != nil {
		return w, err
	}
	out := fmt.Sprintf("%s\n%s", stdout, stderr)
	return w, fmt.Errorf("exit code %d, out: %s", exitCode, out)
}

// Read a file at a given path and returns its content
func (w *Workspace) Read(
	ctx context.Context,
// Path to read the file at
	path string,
) (string, error) {
	return w.Container.File(path).Contents(ctx)
}

// Write a file at a given path with the provided content
func (w *Workspace) Write(
// Path to write the file at
	path string,
// Contents to write
	contents string,
) *Workspace {
	w.Container = w.Container.WithNewFile(path, contents)
	return w
}

// List the available files in tree format
func (w *Workspace) Tree(ctx context.Context) (string, error) {
	return w.Container.WithExec([]string{"tree", "."}).Stdout(ctx)
}
