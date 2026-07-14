// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package archive

import "bytes"

type magicSignature struct {
	magic  []byte
	format Format
}

var signatures = []magicSignature{
	{[]byte{0x1f, 0x8b}, FormatGzip},
	{[]byte("BZh"), FormatBzip2},
	{[]byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00}, FormatXz},
	{[]byte{0x28, 0xb5, 0x2f, 0xfd}, FormatZstd},
	{[]byte{0x04, 0x22, 0x4d, 0x18}, FormatLz4},
	{[]byte{0x50, 0x4b, 0x03, 0x04}, FormatZip},
}

// DetectFormat attempts to detect the compression or archive format of the given data based on its magic bytes.
// If no known magic byte matches, it returns FormatRaw.
func DetectFormat(data []byte) Format {
	for _, sig := range signatures {
		if len(data) >= len(sig.magic) && bytes.HasPrefix(data, sig.magic) {
			return sig.format
		}
	}
	return FormatRaw
}
