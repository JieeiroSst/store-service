package main

import (
	"image"
	"image/color"
	"math"
	"net/http"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
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

	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.POST("/upload", func(c *gin.Context) {
		name := c.PostForm("name")
		email := c.PostForm("email")

		// Multipart form
		form, err := c.MultipartForm()
		if err != nil {
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			return
		}
		files := form.File["files"]

		for _, file := range files {
			filename := filepath.Base(file.Filename)
			if err := c.SaveUploadedFile(file, filename); err != nil {
				c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
				return
			}
		}

		c.String(http.StatusOK, "Uploaded successfully %d files with fields name=%s and email=%s.", len(files), name, email)
	})
	router.Run(":8080")
}
