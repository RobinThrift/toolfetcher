package toolfetcher

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/RobinThrift/toolfetcher/recipes"
	"github.com/RobinThrift/toolfetcher/toolfile"
)

type ToolFetcher struct {
	VersionFile string
	BinDir      string
	StoreDir    string
	Recipes     []recipes.Recipe
}

func (tf *ToolFetcher) Fetch(ctx context.Context, toolname string) error {
	if tf.BinDir == "" {
		tf.BinDir = ".bin"
	}

	if tf.StoreDir == "" {
		tf.StoreDir = path.Join(tf.BinDir, ".store")
	}

	versionFile, err := os.Open(tf.VersionFile)
	if err != nil {
		return fmt.Errorf("error opening version file %s: %w", tf.VersionFile, err)
	}
	defer versionFile.Close()

	var recipe *recipes.Recipe
	for i, r := range tf.Recipes {
		if r.Name == toolname {
			recipe = &tf.Recipes[i]
			break
		}
	}

	if recipe == nil {
		return fmt.Errorf("unknown tool '%s'", toolname)
	}

	entries, err := toolfile.ParseToolFile(versionFile)
	if err != nil {
		return err
	}

	entry, ok := entries[toolname]
	if !ok {
		return fmt.Errorf("unknown tool '%s'", toolname)
	}

	tool := &Tool{
		Name:    toolname,
		Version: entry.Version,
		Recipe:  recipe,
	}

	exists, err := toolSymlinkExists(tool, tf.BinDir, tf.StoreDir)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	inStore, err := toolBinExistsInStore(tool, tf.StoreDir)
	if err != nil {
		return err
	}

	if !inStore {
		err = installTool(ctx, tool, tf.StoreDir)
		if err != nil {
			return err
		}
	}

	err = symlinkTool(tool, tf.BinDir, tf.StoreDir)
	if err != nil {
		return err
	}

	return tool.ExecTest(ctx, tf.BinDir)
}
