package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/RobinThrift/toolfetcher"
	"github.com/RobinThrift/toolfetcher/recipes"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	flags := flag.NewFlagSet("memos-importer", flag.ExitOnError)

	binDir := flags.String("to", "", "bin dir")
	versionfile := flags.String("versionfile", "", "path to version file")

	err := flags.Parse(args)
	if err != nil {
		return fmt.Errorf("invalid usage: %v", err)
	}

	fetcher := toolfetcher.ToolFetcher{
		VersionFile: *versionfile,
		BinDir:      *binDir,
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

	return fetcher.Fetch(ctx, flags.Arg(0))
}
