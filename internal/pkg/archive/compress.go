// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v4"
	"github.com/ulikunitz/xz"
)

// CompressFiles creates an archive containing the given files map (filename -> FileEntry)
// and writes it to the destination io.Writer using the specified format.
func CompressFiles(w io.Writer, format Format, files map[string]FileEntry) error {
	switch format {
	case FormatZip:
		return writeZip(w, files)
	case FormatTar:
		return writeTar(w, files)
	case FormatGzip:
		return writeCompressedTar(w, files, func(tw io.Writer) (io.WriteCloser, error) {
			return gzip.NewWriter(tw), nil
		})
	case FormatZstd:
		return writeCompressedTar(w, files, func(tw io.Writer) (io.WriteCloser, error) {
			return zstd.NewWriter(tw)
		})
	case FormatLz4:
		return writeCompressedTar(w, files, func(tw io.Writer) (io.WriteCloser, error) {
			return lz4.NewWriter(tw), nil
		})
	case FormatXz:
		return writeCompressedTar(w, files, func(tw io.Writer) (io.WriteCloser, error) {
			return xz.NewWriter(tw)
		})
	default:
		// bzip2 compression lacks a pure standard Go implementation for writing.
		// For simplicity, we limit creation to highly utilized and performant formats.
		return fmt.Errorf("%w: %s is not supported for compression yet", ErrUnsupportedFormat, format)
	}
}

func writeZip(w io.Writer, files map[string]FileEntry) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	for name, entry := range files {
		// Use CreateHeader to preserve file mode
		hdr := &zip.FileHeader{
			Name:   name,
			Method: zip.Deflate,
		}
		hdr.SetMode(entry.Mode)

		fw, err := zw.CreateHeader(hdr)
		if err != nil {
			return fmt.Errorf("failed to create zip entry %s: %w", name, err)
		}
		if _, err := io.Copy(fw, bytes.NewReader(entry.Data)); err != nil {
			return fmt.Errorf("failed to write zip entry %s: %w", name, err)
		}
	}
	return nil
}

func writeTar(w io.Writer, files map[string]FileEntry) error {
	tw := tar.NewWriter(w)
	defer tw.Close()

	for name, entry := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: int64(entry.Mode),
			Size: int64(len(entry.Data)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return fmt.Errorf("failed to write tar header for %s: %w", name, err)
		}
		if _, err := io.Copy(tw, bytes.NewReader(entry.Data)); err != nil {
			return fmt.Errorf("failed to write tar entry %s: %w", name, err)
		}
	}
	return nil
}

type compressWriterFactory func(io.Writer) (io.WriteCloser, error)

func writeCompressedTar(w io.Writer, files map[string]FileEntry, factory compressWriterFactory) error {
	cw, err := factory(w)
	if err != nil {
		return fmt.Errorf("failed to initialize compressor: %w", err)
	}
	defer cw.Close()

	return writeTar(cw, files)
}
