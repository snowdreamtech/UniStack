// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractArchive_Errors(t *testing.T) {
	// Invalid zip format (but with zip magic to trigger zip processing)
	zipMagic := []byte{0x50, 0x4b, 0x03, 0x04}
	err := ExtractArchive(zipMagic, t.TempDir())
	assert.Error(t, err)

	// Invalid compressed format (tar inside gzip, but corrupted)
	err = ExtractArchive(append([]byte{0x1f, 0x8b, 0x08}, []byte("invalid")...), t.TempDir())
	assert.Error(t, err)
}

func TestExtractBinary_Errors(t *testing.T) {
	// Zip with invalid data
	zipMagic := []byte{0x50, 0x4b, 0x03, 0x04}
	_, err := ExtractBinary(zipMagic, "file.txt")
	assert.Error(t, err)

	// Gzip with invalid data
	gzMagic := []byte{0x1f, 0x8b, 0x08}
	_, err = ExtractBinary(gzMagic, "file.txt")
	assert.Error(t, err)
}

func TestCompressFiles_Errors(t *testing.T) {
	files := map[string]FileEntry{
		"test": {Data: []byte("a")},
	}

	err := CompressFiles(nil, Format("unsupported"), files)
	assert.Error(t, err)
}

func TestExtractTarFile_MkdirError(t *testing.T) {
	tempDir := t.TempDir()

	// Create a file where a directory needs to be
	err := os.WriteFile(filepath.Join(tempDir, "mydir"), []byte("file"), 0644)
	require.NoError(t, err)

	hdr := &tar.Header{
		Name:     "mydir",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}
	err = extractTarFile(nil, hdr, tempDir)
	assert.Error(t, err) // MkdirAll fails

	hdrFile := &tar.Header{
		Name:     "mydir/file.txt",
		Typeflag: tar.TypeReg,
		Mode:     0644,
	}
	err = extractTarFile(nil, hdrFile, tempDir)
	assert.Error(t, err) // MkdirAll of parent fails
}

func TestExtractZipFile_MkdirError(t *testing.T) {
	tempDir := t.TempDir()

	// Create a file where a directory needs to be
	err := os.WriteFile(filepath.Join(tempDir, "mydir"), []byte("file"), 0644)
	require.NoError(t, err)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	hdrDir := &zip.FileHeader{Name: "mydir/"}
	hdrDir.SetMode(os.ModeDir | 0755)
	_, err = zw.CreateHeader(hdrDir)
	require.NoError(t, err)

	hdrFile := &zip.FileHeader{Name: "mydir/file.txt"}
	hdrFile.SetMode(0644)
	_, err = zw.CreateHeader(hdrFile)
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	err = extractZipFile(zr.File[0], tempDir)
	assert.Error(t, err) // Dir MkdirAll fails

	err = extractZipFile(zr.File[1], tempDir)
	assert.Error(t, err) // File MkdirAll fails
}
