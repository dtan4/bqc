package history

import (
	"fmt"

	"github.com/klauspost/compress/zstd"
)

func compressZstd(in []byte) ([]byte, error) {
	enc, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedBetterCompression))
	if err != nil {
		return []byte{}, fmt.Errorf("create zstd encoder: %w", err)
	}

	return enc.EncodeAll(in, make([]byte, 0, len(in))), nil
}

func decompressZstd(in []byte) ([]byte, error) {
	dec, err := zstd.NewReader(nil)
	if err != nil {
		return []byte{}, fmt.Errorf("create zstd decoder: %w", err)
	}

	v, err := dec.DecodeAll(in, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("decode zstd data: %w", err)
	}

	return v, nil
}
