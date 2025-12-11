// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package manifest

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"google.golang.org/protobuf/encoding/prototext"

	extproctorv1 "zntr.io/extproctor/gen/extproctor/v1"
)

const maxFileSize = 1024 * 1024 // 1MB

// LoadedManifest represents a manifest loaded from a file with its source path.
type LoadedManifest struct {
	*extproctorv1.TestManifest
	SourcePath string
}

// Loader handles loading and parsing of test manifest files.
type Loader struct {
	extensions []string
}

// NewLoader creates a new manifest loader.
func NewLoader() *Loader {
	return &Loader{
		extensions: []string{".textproto", ".prototext", ".txtpb"},
	}
}

// LoadPaths loads manifests from multiple paths (files or directories).
func (l *Loader) LoadPaths(paths []string) ([]*LoadedManifest, error) {
	var manifests []*LoadedManifest

	for _, path := range paths {
		loaded, err := l.LoadPath(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", path, err)
		}
		manifests = append(manifests, loaded...)
	}

	return manifests, nil
}

// LoadPath loads manifests from a single path (file or directory).
func (l *Loader) LoadPath(path string) ([]*LoadedManifest, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		return l.loadDirectory(path)
	}

	manifest, err := l.LoadFile(path)
	if err != nil {
		return nil, err
	}

	return []*LoadedManifest{manifest}, nil
}

// loadDirectory recursively loads all manifest files from a directory.
func (l *Loader) loadDirectory(dir string) ([]*LoadedManifest, error) {
	var manifests []*LoadedManifest

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !l.isManifestFile(path) {
			return nil
		}

		manifest, err := l.LoadFile(path)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", path, err)
		}

		manifests = append(manifests, manifest)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return manifests, nil
}

// LoadFile loads a single manifest file.
func (l *Loader) LoadFile(path string) (*LoadedManifest, error) {
	// Open the file for reading.
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Read the file into a buffer with a maximum size of 1MB to avoid DOS attacks.
	data, err := io.ReadAll(io.LimitReader(f, maxFileSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal the prototext data into a TestManifest message.
	manifest := &extproctorv1.TestManifest{}
	if err := prototext.Unmarshal(data, manifest); err != nil {
		return nil, fmt.Errorf("failed to parse prototext: %w", err)
	}

	// Set default name from filename if not specified.
	if manifest.Name == "" {
		manifest.Name = filepath.Base(path)
	}

	return &LoadedManifest{
		TestManifest: manifest,
		SourcePath:   path,
	}, nil
}

// isManifestFile checks if a file has a recognized manifest extension.
func (l *Loader) isManifestFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return slices.Contains(l.extensions, ext)
}
