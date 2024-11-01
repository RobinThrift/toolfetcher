package toolfetcher

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/RobinThrift/toolfetcher/internal/fs"
	"github.com/RobinThrift/toolfetcher/internal/installer"
	"github.com/RobinThrift/toolfetcher/recipes"
)

type Tool struct {
	Name    string
	Version string
	Recipe  *recipes.Recipe
}

func (t *Tool) VersionedName() string {
	return t.Name + "@" + t.Version
}

func (t *Tool) StoreDir() string {
	return t.Name + "_" + t.Version
}

func (t *Tool) BinPath() string {
	if t.Recipe.Src.BinPath != "" {
		return path.Join(t.StoreDir(), t.Recipe.Src.BinPath)
	}

	if t.Recipe.Src.Type == recipes.SourceTypeBinDownload {
		return path.Join(t.StoreDir(), t.Name)
	}

	return t.StoreDir()
}

func (t *Tool) ExecTest(ctx context.Context, localBinDir string) error {
	if len(t.Recipe.Test) == 0 {
		return nil
	}

	cmd := exec.CommandContext(ctx, path.Join(localBinDir, t.Name), t.Recipe.Test...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running %s %s: %v", t.Name, strings.Join(t.Recipe.Test, " "), err)
	}

	return nil
}

func toolSymlinkExists(t *Tool, binDir string, storeDir string) (bool, error) {
	return fs.SymlinkExists(t.BinPath(), t.StoreDir(), binDir, storeDir)
}

func toolBinExistsInStore(t *Tool, storeDir string) (bool, error) {
	exists, err := fs.FileExists(path.Join(storeDir, t.StoreDir()))
	if err != nil {
		return false, fmt.Errorf("error checking if tool %s has already been downloaded: %v", t.VersionedName(), err)
	}

	return exists, nil
}

func symlinkTool(t *Tool, binDir string, storeDir string) error {
	return fs.Symlink(t.Name, t.BinPath(), binDir, storeDir)
}

func installTool(ctx context.Context, tool *Tool, storeDir string) error {
	switch tool.Recipe.Src.Type {
	case recipes.SourceTypeGoInstall:
		return installer.InstallWithGoInstall(ctx, tool.Recipe, tool.Version, storeDir)
	case recipes.SourceTypeBinDownload:
		return installer.InstallFromBinDownload(ctx, tool.Recipe, tool.Version, storeDir)
	}

	return fmt.Errorf("error installing tool %s: unknown install method %s", tool.VersionedName(), tool.Recipe.Src.Type)
}
