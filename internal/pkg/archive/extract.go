// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v4"
	"github.com/ulikunitz/xz"
)

// NewDecompressReader creates an io.Reader that decompresses the given reader according to the format.
// If the format is not a compression stream (e.g. Zip or Raw), it returns the original reader unchanged.
func NewDecompressReader(r io.Reader, format Format) (io.Reader, error) {
	switch format {
	case FormatGzip:
		gzr, err := gzip.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		// Note: Callers using this must ensure they close the underlying reader if they want gzip to be fully cleaned up.
		// For stream extraction, the reader will just hit EOF.
		return gzr, nil
	case FormatBzip2:
		return bzip2.NewReader(r), nil
	case FormatXz:
		xzr, err := xz.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("failed to create xz reader: %w", err)
		}
		return xzr, nil
	case FormatZstd:
		zstdr, err := zstd.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("failed to create zstd reader: %w", err)
		}
		return zstdr, nil
	case FormatLz4:
		return lz4.NewReader(r), nil
	default:
		// Not a compressed stream (e.g. zip or raw), return as is
		return r, nil
	}
}

// ExtractBinary intelligently extracts a named binary from various archive formats based on magic bytes.
// This is maintained for compatibility and ease of use in simple scenarios.
func ExtractBinary(archiveData []byte, binaryName string) ([]byte, error) {
	format := DetectFormat(archiveData)

	if format == FormatZip {
		zr, err := zip.NewReader(bytes.NewReader(archiveData), int64(len(archiveData)))
		if err != nil {
			return nil, fmt.Errorf("failed to create zip reader: %w", err)
		}
		for _, f := range zr.File {
			if filepath.Base(f.Name) == binaryName && !f.FileInfo().IsDir() {
				rc, err := f.Open()
				if err != nil {
					return nil, fmt.Errorf("failed to open file in zip: %w", err)
				}
				defer rc.Close()
				return io.ReadAll(rc)
			}
		}
		return nil, fmt.Errorf("%w: %s", ErrFileNotFound, binaryName)
	}

	// It's either a compressed tarball, a plain tarball, or raw binary data
	decompressed, err := NewDecompressReader(bytes.NewReader(archiveData), format)
	if err != nil {
		return nil, err
	}

	// If it was a compressed stream or raw data, we now have a reader to its inner contents.
	// We'll peek at it to see if it's a tar archive.
	// Unfortunately, io.Reader doesn't support peaking natively without reading.
	// Since we are working with an in-memory []byte `archiveData`, if it's Raw, it's just the data.
	if format == FormatRaw {
		return archiveData, nil
	}

	// Assume it's a tarball inside the compression stream.
	tr := tar.NewReader(decompressed)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			// If it's not a valid tar, maybe it was a compressed single raw binary?
			// Some formats (like gzip) can compress a single file without a container.
			// Let's reset and read the entire decompressed stream.
			decompressedAgain, _ := NewDecompressReader(bytes.NewReader(archiveData), format)
			return io.ReadAll(decompressedAgain)
		}
		if filepath.Base(hdr.Name) == binaryName && !hdr.FileInfo().IsDir() {
			return io.ReadAll(tr)
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrFileNotFound, binaryName)
}

// ExtractArchive intelligently extracts all contents of an archive into the specified destination directory.
func ExtractArchive(archiveData []byte, destDir string) error {
	format := DetectFormat(archiveData)

	if format == FormatZip {
		zr, err := zip.NewReader(bytes.NewReader(archiveData), int64(len(archiveData)))
		if err != nil {
			return fmt.Errorf("failed to create zip reader: %w", err)
		}
		for _, f := range zr.File {
			if err := extractZipFile(f, destDir); err != nil {
				return err
			}
		}
		return nil
	}

	decompressed, err := NewDecompressReader(bytes.NewReader(archiveData), format)
	if err != nil {
		return err
	}

	if format == FormatRaw {
		// Just write the raw file
		return writeToFile(filepath.Join(destDir, "data.bin"), bytes.NewReader(archiveData), 0644)
	}

	tr := tar.NewReader(decompressed)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			// If not a tar, treat it as a single compressed file
			decompressedAgain, _ := NewDecompressReader(bytes.NewReader(archiveData), format)
			return writeToFile(filepath.Join(destDir, "data.bin"), decompressedAgain, 0644)
		}
		if err := extractTarFile(tr, hdr, destDir); err != nil {
			return err
		}
	}

	return nil
}

func extractZipFile(f *zip.File, destDir string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	path := filepath.Join(destDir, f.Name)
	if f.FileInfo().IsDir() {
		return os.MkdirAll(path, f.Mode())
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return writeToFile(path, rc, f.Mode())
}

func extractTarFile(tr *tar.Reader, hdr *tar.Header, destDir string) error {
	path := filepath.Join(destDir, hdr.Name)
	if hdr.FileInfo().IsDir() {
		return os.MkdirAll(path, os.FileMode(hdr.Mode))
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return writeToFile(path, tr, os.FileMode(hdr.Mode))
}

func writeToFile(path string, r io.Reader, mode os.FileMode) error {
	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, r)
	return err
}
