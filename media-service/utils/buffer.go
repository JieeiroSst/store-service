package utils

import (
	"bytes"
	"io"
	"mime/multipart"
)

func FileHeaderToBytesBuffer(fileHeader *multipart.FileHeader) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	_, err = io.Copy(buf, src)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
