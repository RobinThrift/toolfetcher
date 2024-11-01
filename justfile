local_bin  := absolute_path("./.bin")

go_test_reporter := env("GO_TEST_REPORTER", "pkgname-and-test-fails")


_default:
    @just --list

test *flags="-v -failfast -timeout 15m ./...": (_install-tool "gotestsum")
    {{ local_bin }}/gotestsum --format {{ go_test_reporter }} --format-hide-empty-pkg -- {{ flags }}

test-watch *flags="-v -failfast -timeout 15m ./...": (_install-tool "gotestsum")
    {{ local_bin }}/gotestsum --format {{ go_test_reporter }} --format-hide-empty-pkg --watch -- {{ flags }}

test-report *flags="-v -failfast -timeout 15m ./...": (_install-tool "gotestsum")
    {{ local_bin }}/gotestsum --junitfile "tests.junit.xml" --junitfile-hide-empty-pkg --junitfile-project-name "RobinThrift/toolfetcher" --format {{ go_test_reporter }} --format-hide-empty-pkg -- {{ flags }}

# lint using staticcheck and golangci-lint
lint: (_install-tool "staticcheck") (_install-tool "golangci-lint")
    {{ local_bin }}/staticcheck ./...
    {{ local_bin }}/golangci-lint run ./...

lint-report: (_install-tool "staticcheck") (_install-tool "golangci-lint")
    {{ local_bin }}/golangci-lint run --timeout 5m --out-format=junit-xml ./... > lint.junit.xml
    {{ local_bin }}/staticcheck ./...

fmt:
    @go fmt ./...

clean:
    @rm -rf .bin
    @go clean -cache

# generate a release with the given tag
release tag:
    just changelog {{tag}}
    git add CHANGELOG
    git commit -m "Releasing version {{tag}}"
    git tag {{tag}}
    git push
    git push origin {{tag}}

# generate a changelog using https://github.com/orhun/git-cliff
changelog tag: (_install-tool "git-cliff")
    git-cliff --config CHANGELOG/cliff.toml -o CHANGELOG/CHANGELOG-{{tag}}.md --unreleased --tag {{ tag }} 
    echo "- [CHANGELOG-{{tag}}.md](./CHANGELOG-{{tag}}.md)" >> CHANGELOG/README.md


_install-tool tool:
    @go run ./.scripts/toolfetcher -to {{ local_bin }} -versionfile ./.scripts/TOOL_VERSIONS {{ tool }}
