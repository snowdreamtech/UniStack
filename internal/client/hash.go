// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package client

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

// ValidateHash reads from an io.Reader, computes its SHA-256 hash, and compares it to expectedHash.
// It returns an io.Reader that tees the data to the hash calculator as it's read.
// Actually, it's easier to just compute it while writing to a file.
// CheckHash reads the entire file and checks its hash. But since we want to validate it during or after download,
// we can do a post-download check on the temp file before moving it to its final destination.

// ValidateFileHash calculates the SHA-256 hash of the content in reader and compares it to expectedHash.
func ValidateFileHash(reader io.Reader, expectedHash string) error {
	if expectedHash == "" {
		// If no hash is expected, we implicitly trust it.
		// In a production system, we might want to mandate a hash, but for now we accept empty if not in DB.
		return nil
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return fmt.Errorf("failed to read data for hashing: %w", err)
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if actualHash != expectedHash {
		return fmt.Errorf("hash mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}
