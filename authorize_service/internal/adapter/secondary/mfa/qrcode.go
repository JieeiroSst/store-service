package mfa

import goqrcode "github.com/skip2/go-qrcode"

type QRCode struct{}

func (QRCode) PNG(content string, size int) ([]byte, error) {
	return goqrcode.Encode(content, goqrcode.Medium, size)
}
