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
	"path/filepath"

	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v4"
	"github.com/ulikunitz/xz"
)

// ExtractBinary intelligently extracts a named binary from various archive formats based on magic bytes.
func ExtractBinary(archiveData []byte, binaryName string) ([]byte, error) {
	var binaryData []byte
	var decompressed io.Reader

	// Intelligent extraction based on magic bytes
	if len(archiveData) > 2 && bytes.HasPrefix(archiveData, []byte{0x1f, 0x8b}) {
		// Gzip
		gzr, err := gzip.NewReader(bytes.NewReader(archiveData))
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzr.Close()
		decompressed = gzr
	} else if len(archiveData) > 3 && bytes.HasPrefix(archiveData, []byte("BZh")) {
		// Bzip2
		decompressed = bzip2.NewReader(bytes.NewReader(archiveData))
	} else if len(archiveData) > 5 && bytes.HasPrefix(archiveData, []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00}) {
		// XZ
		xzr, err := xz.NewReader(bytes.NewReader(archiveData))
		if err != nil {
			return nil, fmt.Errorf("failed to create xz reader: %w", err)
		}
		decompressed = xzr
	} else if len(archiveData) > 3 && bytes.HasPrefix(archiveData, []byte{0x28, 0xb5, 0x2f, 0xfd}) {
		// Zstd
		zstdr, err := zstd.NewReader(bytes.NewReader(archiveData))
		if err != nil {
			return nil, fmt.Errorf("failed to create zstd reader: %w", err)
		}
		defer zstdr.Close()
		decompressed = zstdr
	} else if len(archiveData) > 3 && bytes.HasPrefix(archiveData, []byte{0x04, 0x22, 0x4d, 0x18}) {
		// LZ4
		decompressed = lz4.NewReader(bytes.NewReader(archiveData))
	}

	if decompressed != nil {
		tr := tar.NewReader(decompressed)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("failed to read tar archive: %w", err)
			}
			if filepath.Base(hdr.Name) == binaryName && !hdr.FileInfo().IsDir() {
				binaryData, err = io.ReadAll(tr)
				if err != nil {
					return nil, fmt.Errorf("failed to read binary from tar: %w", err)
				}
				break
			}
		}
	} else if len(archiveData) > 4 && bytes.HasPrefix(archiveData, []byte{0x50, 0x4b, 0x03, 0x04}) {
		// Zip
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
				binaryData, err = io.ReadAll(rc)
				rc.Close()
				if err != nil {
					return nil, fmt.Errorf("failed to read binary from zip: %w", err)
				}
				break
			}
		}
	} else {
		// Try treating it as raw binary data just in case
		binaryData = archiveData
	}

	if len(binaryData) == 0 {
		return nil, fmt.Errorf("failed to find %s inside the downloaded archive", binaryName)
	}

	return binaryData, nil
}
