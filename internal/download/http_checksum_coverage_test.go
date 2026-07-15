// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package download

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestVerifyChecksumAutoAllLengths(t *testing.T) {
	// Create a dummy file
	f, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("testdata")
	f.Close()

	d := &HTTPDownloader{}
	ctx := context.Background()

	// length 32 (md5)
	err = d.VerifyChecksum(ctx, f.Name(), strings.Repeat("a", 32))
	if err == nil {
		t.Error("expected checksum mismatch")
	}

	// length 40 (sha1)
	err = d.VerifyChecksum(ctx, f.Name(), strings.Repeat("a", 40))
	if err == nil {
		t.Error("expected checksum mismatch")
	}

	// length 56 (sha224)
	err = d.VerifyChecksum(ctx, f.Name(), strings.Repeat("a", 56))
	if err == nil {
		t.Error("expected checksum mismatch")
	}

	// length 64 (sha256)
	err = d.VerifyChecksum(ctx, f.Name(), strings.Repeat("a", 64))
	if err == nil {
		t.Error("expected checksum mismatch")
	}

	// length 96 (sha384)
	err = d.VerifyChecksum(ctx, f.Name(), strings.Repeat("a", 96))
	if err == nil {
		t.Error("expected checksum mismatch")
	}

	// length 128 (sha512)
	err = d.VerifyChecksum(ctx, f.Name(), strings.Repeat("a", 128))
	if err == nil {
		t.Error("expected checksum mismatch")
	}

	// length invalid
	err = d.VerifyChecksum(ctx, f.Name(), strings.Repeat("a", 10))
	if err == nil {
		t.Error("expected error for invalid length")
	}
}

func TestVerifyChecksumAllAlgorithms(t *testing.T) {
	f, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("testdata")
	f.Close()

	d := &HTTPDownloader{}
	ctx := context.Background()

	algos := []string{
		"md5", "sha1", "sha224", "sha256", "sha384", "sha512",
		"sha3-224", "sha3-256", "sha3-384", "sha3-512",
		"blake2s", "blake2b", "blake3", "unknown",
	}

	for _, algo := range algos {
		// e.g. md5:aaaa...
		err = d.VerifyChecksum(ctx, f.Name(), algo+":"+strings.Repeat("a", 32))
		if err == nil {
			t.Errorf("expected checksum mismatch or unsupported for %s", algo)
		}
	}
}

func TestParseChecksumInvalid(t *testing.T) {
	_, _, err := parseChecksum("")
	if err == nil {
		t.Error("expected error for empty checksum")
	}

	_, _, err = parseChecksum("algo:")
	if err == nil {
		t.Error("expected error for empty hash")
	}

	_, _, err = parseChecksum(":hash")
	if err == nil {
		t.Error("expected error for empty algo")
	}
}
