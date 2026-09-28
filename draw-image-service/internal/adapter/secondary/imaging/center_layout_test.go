package imaging

import (
	"fmt"
	"image"
	"testing"
)

func TestPerimeterSlotsOrder(t *testing.T) {
	want := [][2]float64{
		{0, 0}, {1, 0}, {0, 1}, {1, 1},
		{0.5, 0}, {0.5, 1}, {0, 0.5}, {1, 0.5},
	}
	for n := 0; n <= len(want); n++ {
		slots, _, _ := perimeterSlots(n)
		if len(slots) != n {
			t.Fatalf("n=%d: got %d slots", n, len(slots))
		}
		for i := range slots {
			if slots[i] != want[i] {
				t.Fatalf("n=%d slot %d = %v, want %v", n, i, slots[i], want[i])
			}
		}
	}
}

func TestPerimeterSlotsSpreadEvenlyOnEachEdge(t *testing.T) {
	slots, alongX, alongY := perimeterSlots(4 + 9)
	if alongX != 3 || alongY != 2 {
		t.Fatalf("alongX=%d alongY=%d, want 3 and 2", alongX, alongY)
	}
	var top []float64
	for _, s := range slots[4:] {
		if s[1] == 0 {
			top = append(top, s[0])
		}
	}
	if fmt.Sprint(top) != "[0.25 0.5 0.75]" {
		t.Fatalf("top edge positions %v", top)
	}
}

func TestLayoutCenterPlacesEveryDecorationAroundTheCentre(t *testing.T) {
	canvases := []image.Point{{800, 800}, {1200, 628}, {628, 1200}, {400, 400}}
	centers := []image.Point{{418, 557}, {480, 360}, {1000, 1000}, {1600, 400}, {400, 1600}, {1, 1}}
	decoShapes := []image.Point{{480, 360}, {100, 100}, {60, 200}, {1000, 10}, {1, 1}}

	for _, canvas := range canvases {
		for _, center := range centers {
			for n := 0; n <= 49; n++ {
				decos := make([]image.Point, n)
				for i := range decos {
					decos[i] = decoShapes[i%len(decoShapes)]
				}
				name := fmt.Sprintf("canvas%v/center%v/n=%d", canvas, center, n)
				checkCenterLayout(t, name, canvas, center, decos)
			}
		}
	}
}

func checkCenterLayout(t *testing.T, name string, canvas, center image.Point, decos []image.Point) {
	t.Helper()
	bounds := image.Rect(0, 0, canvas.X, canvas.Y)
	mainRect, rects := layoutCenter(canvas.X, canvas.Y, center, decos)

	if mainRect.Empty() || !mainRect.In(bounds) {
		t.Fatalf("%s: centre %v not inside canvas %v", name, mainRect, bounds)
	}
	if len(rects) != len(decos) {
		t.Fatalf("%s: %d rects for %d decorations", name, len(rects), len(decos))
	}

	maxW := int(float64(canvas.X) * decorationRatio)
	maxH := int(float64(canvas.Y) * decorationRatio)
	mid := image.Pt(mainRect.Min.X+mainRect.Dx()/2, mainRect.Min.Y+mainRect.Dy()/2)
	for i, r := range rects {
		if r.Empty() {
			t.Fatalf("%s: decoration %d was dropped", name, i)
		}
		if !r.In(bounds) {
			t.Fatalf("%s: decoration %d %v outside canvas", name, i, r)
		}
		if r.Dx() > maxW || r.Dy() > maxH {
			t.Fatalf("%s: decoration %d %v larger than %dx%d", name, i, r, maxW, maxH)
		}
		if !r.Inset(-1).Overlaps(mainRect) {
			t.Fatalf("%s: decoration %d %v does not touch the centre %v", name, i, r, mainRect)
		}
		if mainRect.Dx() > 1 && mainRect.Dy() > 1 && mid.In(r) {
			t.Fatalf("%s: decoration %d %v covers the middle of the centre", name, i, r)
		}
		for j := i + 1; j < len(rects); j++ {
			if r.Overlaps(rects[j]) {
				t.Fatalf("%s: decorations %d %v and %d %v overlap", name, i, r, j, rects[j])
			}
		}
	}
}

func TestComposeCenterDrawsEveryDecoration(t *testing.T) {
	p := newTestProcessor(10_000_000)
	for _, n := range []int{1, 4, 5, 8, 9, 13, 20, 49} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			others := make([]image.Image, n)
			sizes := make([]image.Point, n)
			for i := range others {
				others[i] = solid(120, 90, blue)
				sizes[i] = others[i].Bounds().Size()
			}
			center := solid(418, 557, red)
			dst := p.ComposeCenter(center, others, 800, 800)

			_, rects := layoutCenter(800, 800, center.Bounds().Size(), sizes)
			for i, r := range rects {
				c := image.Pt(r.Min.X+r.Dx()/2, r.Min.Y+r.Dy()/2)
				if !isColor(t, dst, c.X, c.Y, blue) {
					t.Fatalf("decoration %d at %v not drawn", i, r)
				}
			}
			if !isColor(t, dst, 400, 400, red) {
				t.Fatal("middle of the centre image must stay visible")
			}
		})
	}
}
