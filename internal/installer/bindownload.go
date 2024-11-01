package installer

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"runtime"
	"text/template"

	"github.com/RobinThrift/toolfetcher/internal/fetch"
	"github.com/RobinThrift/toolfetcher/recipes"
)

func InstallFromBinDownload(ctx context.Context, recipe *recipes.Recipe, version string, storeDir string) error {
	url, err := downloadURLForTool(recipe, version)
	if err != nil {
		return fmt.Errorf("%w: %s@%s: %v", ErrInstalling, recipe.Name, version, err)
	}

	destPath := path.Join(storeDir, recipe.Name+"_"+version)

	err = fetch.DownloadAndUnpackTo(ctx, url, destPath)
	if err != nil {
		return fmt.Errorf("%w: %s@%s: %v", ErrInstalling, recipe.Name, version, err)
	}

	return nil
}

func downloadURLForTool(recipe *recipes.Recipe, version string) (string, error) {
	urlTmpl, err := template.New("").Parse(recipe.Src.URLTemplate)
	if err != nil {
		return "", fmt.Errorf("invalid template: %w", err)
	}

	os := runtime.GOOS
	if recipe.OS != nil {

		if mappedOS, ok := recipe.OS[os]; ok {
			os = mappedOS
		}
	}

	arch := runtime.GOARCH
	if recipe.Arch != nil {
		if mappedArch, ok := recipe.Arch[arch]; ok {
			arch = mappedArch
		}
	}

	var url bytes.Buffer
	err = urlTmpl.Execute(&url, map[string]string{
		"Version": version,
		"OS":      os,
		"Arch":    arch,
	})

	if err != nil {
		return "", fmt.Errorf("error executing URL template: %w", err)
	}

	return url.String(), nil
}
