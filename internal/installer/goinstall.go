package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/RobinThrift/toolfetcher/recipes"
)

func InstallWithGoInstall(ctx context.Context, recipe *recipes.Recipe, version string, storeDir string) error {
	cmd := exec.CommandContext(ctx, "go", "install", recipe.Src.URLTemplate+"@v"+version)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(), "GOBIN="+storeDir)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInstalling, err)
	}

	versionedBinPath := path.Join(storeDir, recipe.Name+"_"+version)

	err = os.Rename(path.Join(storeDir, recipe.Name), versionedBinPath)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInstalling, err)
	}

	return nil
}
