# ToolFetcher

[![CI](https://github.com/RobinThrift/toolfetcher/actions/workflows/ci.yaml/badge.svg)](https://github.com/RobinThrift/toolfetcher/actions/workflows/ci.yaml)
[![BSD-3-Clause license](https://img.shields.io/github/license/RobinThrift/toolfetcher?style=flat-square)](https://github.com/RobinThrift/toolfetcher/blob/main/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/RobinThrift/toolfetcher.svg)](https://pkg.go.dev/github.com/RobinThrift/toolfetcher)
[![Latest Release](https://img.shields.io/github/v/tag/RobinThrift/toolfetcher?sort=semver&style=flat-square)](https://github.com/RobinThrift/toolfetcher/releases/latest)

## Usage

```go
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
		},
	}

	return fetcher.Fetch(ctx, flags.Arg(0))
}
```

## Examples

Example `TOOL_VERSIONS` file:
```
# Go Dev Tools
staticcheck: go://honnef.co/go/tools@0.5.1
golangci-lint: github-releases://golangci/golangci-lint@1.61.0
gotestsum: go://gotest.tools/gotestsum@1.12.0
watchexec: github-releases://watchexec/watchexec@2.2.0

# OpenAPI
oapi-codegen: github-releases://oapi-codegen/oapi-codegen@2.4.1

# SQL Tools
sqlc: github-releases://sqlc-dev/sqlc@1.27.0
tern: github-releases://jackc/tern@2.2.3

# Scanners
syft: github-releases://anchore/syft@1.15.0
grant: github-releases://anchore/grant@0.2.3

# Other Linters
typos: github-releases://crate-ci/typos@1.26.8

# Script Helpers
gum: github-releases://charmbracelet/gum@0.14.5

# Protobuf/GRPC
protoc: github-releases://protocolbuffers/protobuf@27.1
protoc-gen-go: go://google.golang.org/protobuf/@1.35.1
protoc-gen-go-grpc: go://google.golang.org/grpc/cmd/protoc-gen-go-grpc@1.5.1
protoc-gen-connect-go: go://connectrpc.com/connect/@1.17.0
```


Example Renovate Bot config:
```json
{
    "enabledManagers": ["custom.regex"],
    "customManagers": [
        {
            "customType": "regex",
            "fileMatch": ["^.tools/TOOL_VERSIONS$"],
            "matchStrings": [
                "(?<depName>.+?): *(?<datasource>github-releases|go)://(?<packageName>.+?)@(?<currentValue>[\\d\\.]+)"
            ],
            "versioningTemplate": "semver"
        }
    ]
}
```

## License

BSD-3-Clause license
