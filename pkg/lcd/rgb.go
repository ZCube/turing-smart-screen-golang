package lcd

import (
	"image"

	"github.com/synthread/framebuffer"
)

func NewRGB565(r image.Rectangle) *framebuffer.RGB565 {
	return &framebuffer.RGB565{
		Pix:    make([]uint8, 2*r.Dx()*r.Dy()),
		Stride: 2 * r.Dx(),
		Rect:   r,
	}
}
