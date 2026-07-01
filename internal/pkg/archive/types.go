// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package archive

import (
	"fmt"
	"os"
	"time"
)

// Format represents an archive or compression format.
type Format string

const (
	FormatUnknown Format = "unknown"
	FormatRaw     Format = "raw"
	FormatZip     Format = "zip"
	FormatTar     Format = "tar" // plain tar
	FormatGzip    Format = "gzip"
	FormatBzip2   Format = "bzip2"
	FormatXz      Format = "xz"
	FormatZstd    Format = "zstd"
	FormatLz4     Format = "lz4"
)

// Extension returns the typical file extension for the format.
func (f Format) Extension() string {
	switch f {
	case FormatZip:
		return ".zip"
	case FormatTar:
		return ".tar"
	case FormatGzip:
		return ".gz"
	case FormatBzip2:
		return ".bz2"
	case FormatXz:
		return ".xz"
	case FormatZstd:
		return ".zst"
	case FormatLz4:
		return ".lz4"
	default:
		return ""
	}
}

// String returns the string representation of the format.
func (f Format) String() string {
	return string(f)
}

// IsCompression returns true if the format is purely a compression stream (not a container like zip or tar).
func (f Format) IsCompression() bool {
	switch f {
	case FormatGzip, FormatBzip2, FormatXz, FormatZstd, FormatLz4:
		return true
	default:
		return false
	}
}

// IsContainer returns true if the format is an archive container (like zip or tar).
func (f Format) IsContainer() bool {
	switch f {
	case FormatZip, FormatTar:
		return true
	default:
		return false
	}
}

// Error definitions
var (
	ErrUnsupportedFormat = fmt.Errorf("unsupported archive format")
	ErrFileNotFound      = fmt.Errorf("file not found in archive")
	ErrInvalidData       = fmt.Errorf("invalid archive data")
)

// FileEntry represents a file to be added to an archive.
type FileEntry struct {
	Data    []byte
	Mode    os.FileMode
	ModTime time.Time

	// Tar specific metadata
	Uid        int
	Gid        int
	Uname      string
	Gname      string
	PAXRecords map[string]string // Custom key-value extensions

	// Zip specific metadata
	Comment string
}
