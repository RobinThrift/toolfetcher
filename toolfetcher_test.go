package toolfetcher

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/RobinThrift/toolfetcher/recipes"
	"github.com/RobinThrift/toolfetcher/toolfile"
	"github.com/stretchr/testify/require"
)

func TestToolFetcher_Fetch(t *testing.T) {
	cwd := t.TempDir()

	toolfilePath := path.Join(cwd, "TOOL_VERSIONS")
	toolfileContent, err := os.ReadFile(".scripts/TOOL_VERSIONS")
	if err != nil {
		require.NoError(t, err)
	}

	err = os.WriteFile(toolfilePath, toolfileContent, 0o644)
	if err != nil {
		require.NoError(t, err)
	}

	toolnames := readToolFile(t, toolfilePath)

	fetcher := ToolFetcher{
		VersionFile: toolfilePath,
		BinDir:      cwd,
		Recipes: []recipes.Recipe{
			{
				Name: "staticcheck",
				Src: recipes.Source{
					Type:        recipes.SourceTypeGoInstall,
					URLTemplate: "honnef.co/go/tools/cmd/staticcheck",
				},
				Test: []string{"--version"},
			},

			{
				Name: "golangci-lint",
				Src: recipes.Source{
					Type:        recipes.SourceTypeGoInstall,
					URLTemplate: "github.com/golangci/golangci-lint/cmd/golangci-lint",
				},
				Test: []string{"--version"},
			},

			{
				Name: "gotestsum",
				Src: recipes.Source{
					Type:        recipes.SourceTypeGoInstall,
					URLTemplate: "gotest.tools/gotestsum",
				},
				Test: []string{"--version"},
			},

			{
				Name: "git-cliff",
				Src: recipes.Source{
					Type:        recipes.SourceTypeBinDownload,
					URLTemplate: "https://github.com/orhun/git-cliff/releases/download/v{{ .Version }}/git-cliff-{{ .Version }}-{{ .Arch }}-{{ .OS }}.tar.gz",
				},
				OS:   map[string]string{"darwin": "apple-darwin", "linux": "unknown-linux-gnu"},
				Arch: map[string]string{"arm64": "aarch64", "amd64": "x86_64"},
				Test: []string{"--version"},
			},
		},
	}

	for _, tool := range toolnames {
		t.Run(tool, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			err := fetcher.Fetch(ctx, tool)
			if err != nil {
				require.NoError(t, err)
			}
		})
	}

}

func readToolFile(t *testing.T, toolfilePath string) []string {
	file, err := os.Open(toolfilePath)
	if err != nil {
		require.NoError(t, err)
	}
	defer file.Close()

	entries, err := toolfile.ParseToolFile(file)
	if err != nil {
		require.NoError(t, err)
	}

	toolnames := make([]string, 0, len(entries))
	for name := range entries {
		toolnames = append(toolnames, name)
	}

	return toolnames
}
