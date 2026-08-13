// Package qrcode implements ports.QRCodeGenerator.
package qrcode

import (
	"fmt"
	"strings"

	qr "github.com/skip2/go-qrcode"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) PNG(data string, size int, darkColor, lightColor string) ([]byte, error) {
	code, err := qr.New(data, qr.Medium)
	if err != nil {
		return nil, err
	}
	dark, err := parseHexColor(darkColor)
	if err != nil {
		return nil, err
	}
	light, err := parseHexColor(lightColor)
	if err != nil {
		return nil, err
	}
	code.ForegroundColor = dark
	code.BackgroundColor = light
	return code.PNG(size)
}

func (g *Generator) SVG(data string, size int, darkColor, lightColor string) (string, error) {
	code, err := qr.New(data, qr.Medium)
	if err != nil {
		return "", err
	}
	bitmap := code.Bitmap()
	n := len(bitmap)
	if n == 0 {
		return "", fmt.Errorf("qrcode: empty bitmap")
	}
	// go-qrcode's Bitmap() already includes its own quiet-zone margin.
	moduleSize := float64(size) / float64(n)

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" shape-rendering="crispEdges">`, size, size, size, size)
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="%s"/>`, size, size, lightColor)
	b.WriteString(`<path fill="` + darkColor + `" d="`)
	for y, row := range bitmap {
		for x, dark := range row {
			if !dark {
				continue
			}
			px := float64(x) * moduleSize
			py := float64(y) * moduleSize
			fmt.Fprintf(&b, "M%.3f %.3fh%.3fv%.3fh-%.3fz", px, py, moduleSize, moduleSize, moduleSize)
		}
	}
	b.WriteString(`"/></svg>`)
	return b.String(), nil
}

func parseHexColor(s string) (c colorRGBA, err error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return c, fmt.Errorf("qrcode: invalid hex color %q", s)
	}
	var r, g, bl uint8
	if _, err := fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &bl); err != nil {
		return c, fmt.Errorf("qrcode: invalid hex color %q", s)
	}
	return colorRGBA{r, g, bl, 0xff}, nil
}

type colorRGBA struct{ R, G, B, A uint8 }

func (c colorRGBA) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R)
	r |= r << 8
	g = uint32(c.G)
	g |= g << 8
	b = uint32(c.B)
	b |= b << 8
	a = uint32(c.A)
	a |= a << 8
	return
}
