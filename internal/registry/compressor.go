package registry

import (
	"fmt"
	"io"
	"os"

	"github.com/klauspost/compress/zstd"
)

// CompressZstd reads the source file and writes a .zst compressed version
func CompressZstd(sourcePath, destPath string) error {
	inFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file for compression: %w", err)
	}
	defer inFile.Close()

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer outFile.Close()

	// Use DefaultCompression (which is level 3, a good balance of speed/size)
	enc, err := zstd.NewWriter(outFile, zstd.WithEncoderLevel(zstd.SpeedDefault))
	if err != nil {
		return fmt.Errorf("failed to initialize zstd encoder: %w", err)
	}
	defer enc.Close()

	if _, err := io.Copy(enc, inFile); err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}

	return nil
}
