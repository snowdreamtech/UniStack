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
)

// CompressFiles creates an archive containing the given files map (filename -> content)
// and writes it to the destination io.Writer using the specified format.
func CompressFiles(w io.Writer, format Format, files map[string][]byte) error {
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
	default:
		// bzip2 and xz compression are complex or lacking pure standard Go implementations for writing.
		// For simplicity, we limit creation to highly utilized and performant formats.
		return fmt.Errorf("%w: %s is not supported for compression yet", ErrUnsupportedFormat, format)
	}
}

func writeZip(w io.Writer, files map[string][]byte) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	for name, data := range files {
		fw, err := zw.Create(name)
		if err != nil {
			return fmt.Errorf("failed to create zip entry %s: %w", name, err)
		}
		if _, err := io.Copy(fw, bytes.NewReader(data)); err != nil {
			return fmt.Errorf("failed to write zip entry %s: %w", name, err)
		}
	}
	return nil
}

func writeTar(w io.Writer, files map[string][]byte) error {
	tw := tar.NewWriter(w)
	defer tw.Close()

	for name, data := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0600,
			Size: int64(len(data)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return fmt.Errorf("failed to write tar header for %s: %w", name, err)
		}
		if _, err := io.Copy(tw, bytes.NewReader(data)); err != nil {
			return fmt.Errorf("failed to write tar entry %s: %w", name, err)
		}
	}
	return nil
}

type compressWriterFactory func(io.Writer) (io.WriteCloser, error)

func writeCompressedTar(w io.Writer, files map[string][]byte, factory compressWriterFactory) error {
	cw, err := factory(w)
	if err != nil {
		return fmt.Errorf("failed to initialize compressor: %w", err)
	}
	defer cw.Close()

	return writeTar(cw, files)
}
