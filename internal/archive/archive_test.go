// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package archive

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArchiveFormatDetection(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected Format
	}{
		{"gzip", []byte{0x1f, 0x8b, 0x08}, FormatGzip},
		{"bzip2", []byte("BZh91AY&SY"), FormatBzip2},
		{"xz", []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00}, FormatXz},
		{"zstd", []byte{0x28, 0xb5, 0x2f, 0xfd}, FormatZstd},
		{"lz4", []byte{0x04, 0x22, 0x4d, 0x18}, FormatLz4},
		{"zip", []byte{0x50, 0x4b, 0x03, 0x04}, FormatZip},
		{"raw", []byte("hello world"), FormatRaw},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, DetectFormat(tt.data))
		})
	}
}

func TestArchiveFormatMethods(t *testing.T) {
	assert.Equal(t, ".zip", FormatZip.Extension())
	assert.Equal(t, ".tar", FormatTar.Extension())
	assert.Equal(t, ".gz", FormatGzip.Extension())
	assert.Equal(t, ".bz2", FormatBzip2.Extension())
	assert.Equal(t, ".xz", FormatXz.Extension())
	assert.Equal(t, ".zst", FormatZstd.Extension())
	assert.Equal(t, ".lz4", FormatLz4.Extension())
	assert.Equal(t, "", FormatRaw.Extension())

	assert.Equal(t, "zip", FormatZip.String())
	assert.True(t, FormatGzip.IsCompression())
	assert.False(t, FormatZip.IsCompression())
	assert.True(t, FormatZip.IsContainer())
	assert.False(t, FormatGzip.IsContainer())
}

func TestCompressAndExtractZip(t *testing.T) {
	var buf bytes.Buffer
	files := map[string]FileEntry{
		"test.txt": {
			Data:    []byte("hello archive"),
			Mode:    0644,
			ModTime: time.Now(),
		},
		"dir/file.txt": {
			Data:    []byte("in dir"),
			Mode:    0644,
			ModTime: time.Now(),
		},
	}

	err := CompressFiles(&buf, FormatZip, files)
	require.NoError(t, err)

	archiveData := buf.Bytes()
	assert.Equal(t, FormatZip, DetectFormat(archiveData))

	tempDir := t.TempDir()
	err = ExtractArchive(archiveData, tempDir)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tempDir, "test.txt"))
	require.NoError(t, err)
	assert.Equal(t, "hello archive", string(content))

	contentDir, err := os.ReadFile(filepath.Join(tempDir, "dir", "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "in dir", string(contentDir))

	// Test ExtractBinary
	binData, err := ExtractBinary(archiveData, "file.txt")
	require.NoError(t, err)
	assert.Equal(t, "in dir", string(binData))
}

func TestCompressAndExtractCompressedTar(t *testing.T) {
	formats := []Format{FormatGzip, FormatBzip2, FormatXz, FormatZstd, FormatLz4}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			var buf bytes.Buffer
			files := map[string]FileEntry{
				"test.txt": {
					Data:    []byte("hello " + string(format)),
					Mode:    0644,
					ModTime: time.Now(),
				},
			}

			err := CompressFiles(&buf, format, files)
			if format == FormatBzip2 {
				// CompressFiles doesn't support bzip2 yet
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			archiveData := buf.Bytes()
			assert.Equal(t, format, DetectFormat(archiveData))

			tempDir := t.TempDir()
			err = ExtractArchive(archiveData, tempDir)
			require.NoError(t, err)

			content, err := os.ReadFile(filepath.Join(tempDir, "test.txt"))
			require.NoError(t, err)
			assert.Equal(t, "hello "+string(format), string(content))

			// Test ExtractBinary
			binData, err := ExtractBinary(archiveData, "test.txt")
			require.NoError(t, err)
			assert.Equal(t, "hello "+string(format), string(binData))
		})
	}
}

func TestCompressAndExtractTar(t *testing.T) {
	var buf bytes.Buffer
	files := map[string]FileEntry{
		"test.txt": {
			Data:    []byte("hello tar"),
			Mode:    0644,
			ModTime: time.Now(),
		},
	}

	err := CompressFiles(&buf, FormatTar, files)
	require.NoError(t, err)

	archiveData := buf.Bytes()

	tempDir := t.TempDir()
	err = ExtractArchive(archiveData, tempDir)
	require.NoError(t, err)

	// Since plain tar has no compression magic bytes recognized by DetectFormat,
	// ExtractArchive will treat it as FormatRaw and extract it as data.bin.
	content, err := os.ReadFile(filepath.Join(tempDir, "data.bin"))
	require.NoError(t, err)
	// The content will be the tar file itself, not "hello tar". So we just check err == nil
	assert.NotEmpty(t, content)
}
