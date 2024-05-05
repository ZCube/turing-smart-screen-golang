package lcd

import (
	"errors"
	"image"

	"github.com/zcube/turing-smart-screen-golang/pkg/common"
	"github.com/zcube/turing-smart-screen-golang/pkg/lcd/rev_a"
)

type LcdComm interface {
	Width() int
	Height() int
	Revision() string
	Close() error
	WriteData(byteBuffer []byte) (int, error)
	ReadData(readSize int) ([]byte, error)
	Reset() error
	Clear() error
	ScreenOff() error
	ScreenOn() error
	SetBrightness(level int) error
	SetOrientation(orientation common.Orientation) error
	DisplayImage(imageRGB656LE []byte, x, y, imageWidth, imageHeight int) error
}

func New(deviceName string, revision common.Revision) (LcdComm, error) {
	switch revision {
	case common.RevisionA:
		return rev_a.NewLcdCommA(deviceName)
	default:
		err := errors.New("Unsupported revision")
		return nil, err
	}
}

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func DisplayImage(lcdComm LcdComm, img image.Image, x, y, imageWidth, imageHeight int) error {
	width := lcdComm.Width()
	height := lcdComm.Height()

	if imageHeight == 0 {
		imageHeight = img.Bounds().Dy()
	}
	if imageWidth == 0 {
		imageWidth = img.Bounds().Dx()
	}

	if x > width {
		err := errors.New("Image X coordinate must be <= display width")
		return err
	}
	if y > height {
		err := errors.New("Image Y coordinate must be <= display height")
		return err
	}
	if imageHeight <= 0 {
		err := errors.New("Image height must be > 0")
		return err
	}
	if imageWidth <= 0 {
		err := errors.New("Image width must be > 0")
		return err
	}

	if x+imageWidth > width {
		imageWidth = width - x
	}
	if y+imageHeight > height {
		imageHeight = height - y
	}

	if imageWidth != img.Bounds().Dx() || imageHeight != img.Bounds().Dy() {
		cropSize := image.Rect(0, 0, width, height)
		paddingSize := image.Point{x, y}
		cropSize = cropSize.Add(paddingSize)
		img = img.(SubImager).SubImage(cropSize)
	}

	rgb565 := NewRGB565(image.Rect(0, 0, imageWidth, imageHeight))
	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			rgb565.Set(x, y, img.At(x, y))
		}
	}

	x0, y0 := x, y

	err := lcdComm.DisplayImage(rgb565.Pix, x0, y0, imageWidth, imageHeight)
	if err != nil {
		return err
	}
	return nil
}
