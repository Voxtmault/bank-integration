package bank_integration_utils

import (
	"bytes"
	"compress/gzip"
	"io"
)

func CompressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecompressData(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(data)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	uncompressedData, err := io.ReadAll(gz)
	if err != nil {
		return nil, err
	}
	return uncompressedData, nil
}
