package toolfile

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

var ErrMissingName = errors.New("missing tool name")
var ErrMissingVersion = errors.New("missing tool version")

type Entries map[string]Entry

type Entry struct {
	Name    string
	Version string
}

func ParseToolFile(r io.Reader) (Entries, error) {
	scanner := bufio.NewScanner(r)

	entries := Entries{}

	lineNum := 1
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		entry, err := parseLine(line, lineNum)
		if err != nil {
			return nil, err
		}

		lineNum++
		entries[entry.Name] = entry
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading version file: %w", err)
	}

	return entries, nil
}

func parseLine(line []byte, lineNum int) (Entry, error) {
	entry := Entry{}

	firstColonIndex := bytes.Index(line, []byte(":"))
	if firstColonIndex == -1 {
		return entry, fmt.Errorf("%w on line %d", ErrMissingName, lineNum)
	}

	if firstColonIndex+1 >= len(line) {
		return entry, fmt.Errorf("%w on line %d: expected source and version number, got EOL", ErrMissingVersion, lineNum)
	}

	if line[firstColonIndex+1] == '/' {
		return entry, fmt.Errorf("%w on line %d", ErrMissingName, lineNum)
	}

	entry.Name = string(line[0:firstColonIndex])

	versionMarkerIndex := bytes.Index(line, []byte("@"))
	if versionMarkerIndex == -1 {
		return entry, fmt.Errorf("%w for tool %s on line %d: no version marker '@' found", ErrMissingVersion, entry.Name, lineNum)
	}

	if versionMarkerIndex+1 >= len(line) {
		return entry, fmt.Errorf("%w for tool %s on line %d: expected version number, got EOL", ErrMissingVersion, entry.Name, lineNum)
	}

	entry.Version = string(line[versionMarkerIndex+1:])
	if entry.Version == "" {
		return entry, fmt.Errorf("%w for tool %s on line %d: no version found", ErrMissingVersion, entry.Name, lineNum)
	}

	return entry, nil
}
