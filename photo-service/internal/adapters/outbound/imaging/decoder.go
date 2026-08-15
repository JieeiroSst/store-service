package imaging

import (
	"bytes"
	"context"
	"fmt"
	"image"

	"github.com/disintegration/imaging"
	"github.com/gen2brain/webp"

	"github.com/JIeeiroSst/photo-service/internal/application/ports"
	"github.com/JIeeiroSst/photo-service/internal/domain"
)

type Decoder struct{}

func NewDecoder() *Decoder {
	return &Decoder{}
}

var _ ports.ImageDecoder = (*Decoder)(nil)

func (d *Decoder) Decode(ctx context.Context, data []byte) (image.Image, error) {
	if isWebP(data) {
		img, err := webp.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("%w: %v", domain.ErrInvalidImageData, err)
		}
		return img, nil
	}

	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidImageData, err)
	}
	return img, nil
}

func isWebP(data []byte) bool {
	return len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP"
}
