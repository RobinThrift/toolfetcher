package fetch

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
)

func DownloadAndUnpackTo(ctx context.Context, url string, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("error creating new request for URL '%s': %w", url, err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error fetching resource from '%s': %w", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error fetching resource from '%s': %v %v", url, res.StatusCode, res.Status)
	}

	ext := path.Ext(url)
	tempFilePattern := path.Base(destPath) + "-*" + ext

	tmpFile, err := os.CreateTemp("", tempFilePattern)
	if err != nil {
		return fmt.Errorf("error creating temporary file")
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	_, err = io.Copy(tmpFile, res.Body)
	if err != nil {
		return fmt.Errorf("error downloading file: %w", err)
	}

	err = unpackArchive(tmpFile.Name(), destPath, ext)
	if err != nil {
		return err
	}

	return nil
}

func unpackArchive(filePath string, destPath string, ext string) error {
	switch ext {
	case ".zip":
		return unpackZipArchive(filePath, destPath)
	case ".tar":
		return unpackTarArchive(filePath, destPath)
	case ".gz":
		return unpackTarGzipArchive(filePath, destPath)
	case ".xz":
		return unpackTarXZArchive(filePath, destPath)
	}

	return fmt.Errorf("unknown archive %s", ext)
}

func unpackZipArchive(filePath string, destPath string) error {
	archive, err := zip.OpenReader(filePath)
	if err != nil {
		return fmt.Errorf("error constructing new ZIP reader: %w", err)
	}

	err = os.MkdirAll(destPath, 0o755)
	if err != nil {
		return fmt.Errorf("error creating directory %s: %w", destPath, err)
	}

	for _, compressed := range archive.File {
		f, err := compressed.Open()
		if err != nil {
			return fmt.Errorf("error openening compressed file %s: %w", compressed.Name, err)
		}
		defer f.Close()

		destFilePath := path.Join(destPath, compressed.Name)

		err = os.MkdirAll(path.Dir(destFilePath), 0o755)
		if err != nil {
			return fmt.Errorf("error creating directory %s: %w", path.Dir(destFilePath), err)
		}

		if compressed.Mode().IsDir() {
			err = os.Mkdir(destFilePath, 0o755)
			if err != nil {
				return fmt.Errorf("error creating directory %s: %w", destFilePath, err)
			}
		} else {
			destFile, err := os.OpenFile(destFilePath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, compressed.Mode())
			if err != nil {
				return fmt.Errorf("error creating file %s: %w", destFilePath, err)
			}
			defer destFile.Close()

			_, err = io.Copy(destFile, f)
			if err != nil {
				return fmt.Errorf("error decompressing file %s to %s: %w", compressed.Name, destFilePath, err)
			}
		}
	}

	return nil
}

func unpackTarGzipArchive(filePath string, destPath string) error {
	r, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer r.Close()

	zr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("error constructing new gzip reader: %w", err)
	}
	defer zr.Close()

	return unpackTarReader(zr, destPath)
}

func unpackTarArchive(filePath string, destPath string) error {
	r, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer r.Close()

	return unpackTarReader(r, destPath)
}

func unpackTarReader(r io.Reader, destPath string) error {
	archive := tar.NewReader(r)

	err := os.MkdirAll(destPath, 0o755)
	if err != nil {
		return fmt.Errorf("error creating directory %s: %w", destPath, err)
	}

	for {
		header, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		destFilePath := path.Join(destPath, header.FileInfo().Name())
		if header.FileInfo().IsDir() {
			err = os.MkdirAll(destFilePath, 0o755)
			if err != nil {
				return fmt.Errorf("error creating directory %s: %w", destFilePath, err)
			}
		} else {
			destFile, err := os.OpenFile(destFilePath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, header.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("error creating file %s: %w", destFilePath, err)
			}
			defer destFile.Close()

			_, err = io.Copy(destFile, archive)
			if err != nil {
				return fmt.Errorf("error decompressing file %s to %s: %w", header.Name, destFilePath, err)
			}
		}
	}

	return nil
}

func unpackTarXZArchive(filePath string, destPath string) error {
	tmpdir, err := os.MkdirTemp("", path.Base(destPath))
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmpdir) // ignore errors
	}()

	cmd := exec.Command("tar", "--strip-components=1", "-C", tmpdir, "-xf", filePath)
	err = cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		var output string
		if errors.As(err, &exitErr) {
			output = "\n" + string(exitErr.Stderr)
		}

		return fmt.Errorf("error unpacking .xz archive using the 'tar' command: %w%s", err, output)
	}

	return os.Rename(tmpdir, destPath)
}
