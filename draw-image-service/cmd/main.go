package main

import (
	"image"
	"image/color"
	"math"

	"github.com/disintegration/imaging"
)

func main() {
	// input files
	files := []string{"01.jpeg", "02.jpeg", "03.jpeg", "04.jpeg"}

	// load images and make 100x100 thumbnails of them
	sideWidth, sideHeight := 100, 100
	var thumbnails []image.Image
	for _, f := range files {
		img, err := imaging.Open(f)
		if err != nil {
			panic(err)
		}
		thumb := imaging.Thumbnail(img, sideWidth, sideHeight, imaging.CatmullRom)
		thumbnails = append(thumbnails, thumb)
	}

	// create a new blank image
	dst := imaging.New(sideWidth*len(thumbnails), sideHeight, color.NRGBA{0, 0, 0, 0})

	// paste thumbnails into the new image side by side
	for i, thumb := range thumbnails {
		dst = imaging.Paste(dst, thumb, image.Pt(i*sideWidth, 0))
	}

	// save the combined image to file
	if err := imaging.Save(dst, "dst.jpg"); err != nil {
		panic(err)
	}
	thumbLen := len(thumbnails)
	columns := 2
	
	rows := int(math.Ceil(float64(thumbLen) / float64(columns)))

	dst = imaging.New(sideWidth*columns, sideHeight*rows, color.NRGBA{0, 0, 0, 0})
	for i, thumb := range thumbnails {
		pX := sideWidth * (i % 2)
		pY := sideHeight * (i / columns)
		dst = imaging.Paste(dst, thumb, image.Pt(pX, pY))
	}

	if err := imaging.Save(dst, "dst2.jpg", imaging.JPEGQuality(95)); err != nil {
		panic(err)
	}
}
