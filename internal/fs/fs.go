package fs

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

func FileExists(file string) (bool, error) {
	if file == "" {
		return false, nil
	}

	_, err := os.Stat(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, fmt.Errorf("error checking if file '%s' exists: %v", file, err)
	}

	return true, nil
}

func SymlinkExists(name string, versionedBin string, targetDir string, srcDir string) (bool, error) {
	versionedBinPath := path.Join(srcDir, versionedBin)
	localFinalBinPath := path.Join(targetDir, name)

	symlinked, err := filepath.EvalSymlinks(localFinalBinPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("error checking if tool %s has already been linked: %v", name, err)
	}

	return symlinked == versionedBinPath, nil
}

func Symlink(name string, versionedBin string, targetDir string, srcDir string) error {
	versionedBinPath := path.Join(srcDir, versionedBin)
	localFinalBinPath := path.Join(targetDir, name)

	err := os.Remove(localFinalBinPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("error symlinking '%s' to '%s': can't remove existing file at target: %v", versionedBinPath, localFinalBinPath, err)
		}
	}

	err = os.MkdirAll(path.Dir(localFinalBinPath), 0o755)
	if err != nil {
		return fmt.Errorf("error symlinking '%s' to '%s': error creating target directory: %v", versionedBinPath, localFinalBinPath, err)

	}

	err = os.Symlink(versionedBinPath, localFinalBinPath)
	if err != nil {
		return fmt.Errorf("error symlinking '%s' to '%s': %v", versionedBinPath, localFinalBinPath, err)
	}

	return nil
}
