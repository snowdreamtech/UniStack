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

func TestExtractTarFile(t *testing.T) {
	tempDir := t.TempDir()

	// Test Directory
	hdr := &tar.Header{
		Name:     "mydir",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}
	err := extractTarFile(nil, hdr, tempDir)
	require.NoError(t, err)
	stat, err := os.Stat(filepath.Join(tempDir, "mydir"))
	require.NoError(t, err)
	assert.True(t, stat.IsDir())

	// Test Symlink
	hdrSymlink := &tar.Header{
		Name:     "symlink",
		Typeflag: tar.TypeSymlink,
		Linkname: "mydir",
	}
	err = extractTarFile(nil, hdrSymlink, tempDir)
	require.NoError(t, err)

	// Test Hardlink (Link)
	hdrLink := &tar.Header{
		Name:     "hardlink",
		Typeflag: tar.TypeLink,
		Linkname: "mydir",
	}
	// Let's test hardlinking a regular file instead
	hdrFile := &tar.Header{
		Name:     "file.txt",
		Typeflag: tar.TypeReg,
		Size:     0,
	}
	err = extractTarFile(tar.NewReader(bytes.NewReader(nil)), hdrFile, tempDir)
	require.NoError(t, err)

	hdrLink.Linkname = "file.txt"
	err = extractTarFile(nil, hdrLink, tempDir)
	require.NoError(t, err)
}

func TestExtractZipFileModes(t *testing.T) {
	tempDir := t.TempDir()

	// Create an in-memory zip containing a directory and a symlink
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Add directory
	hdrDir := &zip.FileHeader{
		Name: "mydir/",
	}
	hdrDir.SetMode(os.ModeDir | 0755)
	_, err := zw.CreateHeader(hdrDir)
	require.NoError(t, err)

	// Add symlink
	hdrSymlink := &zip.FileHeader{
		Name: "symlink",
	}
	hdrSymlink.SetMode(os.ModeSymlink | 0777)
	fw, err := zw.CreateHeader(hdrSymlink)
	require.NoError(t, err)
	fw.Write([]byte("mydir"))

	require.NoError(t, zw.Close())

	// Read it back
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	for _, f := range zr.File {
		err = extractZipFile(f, tempDir)
		require.NoError(t, err)
	}

	stat, err := os.Stat(filepath.Join(tempDir, "mydir"))
	require.NoError(t, err)
	assert.True(t, stat.IsDir())
}
