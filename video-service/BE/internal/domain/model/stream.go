package model

import "io"

type VideoStream struct {
	Video   Video
	Content io.ReadSeeker
}

type Asset struct {
	Data        []byte
	ContentType string
}
