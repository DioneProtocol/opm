package repository

import (
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/ava-labs/apm/types"
)

// Metadata represents a repository, its source, and the last synced commit.
type Metadata struct {
	Alias  string        `yaml:"alias"`
	URL    string        `yaml:"url"`
	Commit plumbing.Hash `yaml:"commit"`
}

// Registry is a list of repositories that support a single plugin alias.
// e.g. foo/plugin, bar/plugin => plugin: [foo, bar]
type Registry struct {
	Repositories []string `yaml:"repositories"`
}

// Plugin stores a plugin definition alongside the plugin-repository's commit
// it was downloaded from.
// TODO gc plugins
type Plugin[T types.Plugin] struct {
	Plugin T             `yaml:"plugin"`
	Commit plumbing.Hash `yaml:"commit"`
}
