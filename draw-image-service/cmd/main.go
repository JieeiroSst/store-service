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
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.POST("/upload", func(c *gin.Context) {
		// Multipart form
		form, err := c.MultipartForm()
		if err != nil {
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			return
		}
		images := form.File["files"]

		files := make([]string, 0)

		for _, image := range images {
			filename := filepath.Base(image.Filename)
			// if err := c.SaveUploadedFile(image, filename); err != nil {
			// 	c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
			// 	return
			// }
			files = append(files, filename)
		}

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

		c.String(http.StatusOK, "Uploaded successfully %d files.", len(files))
	})
	router.Run(":8080")
}
