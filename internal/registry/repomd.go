// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package registry

// RepoMd represents the metadata for the package registry.
// It is typically serialized to repomd.json.
type RepoMd struct {
	Timestamp int64  `json:"timestamp"`
	Hash      string `json:"hash"` // sha256:abcd...
	Path      string `json:"path"` // e.g., repodata/packages.db.zst
}
