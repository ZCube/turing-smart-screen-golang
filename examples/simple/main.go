package main

import (
	"fmt"
	"log"
	"math/rand"

	"image/color"
	_ "image/jpeg"
	_ "image/png"

	"github.com/fogleman/gg"
	"github.com/zcube/turing-smart-screen-golang/pkg/common"
	"github.com/zcube/turing-smart-screen-golang/pkg/lcd"
	"go.bug.st/serial/enumerator"
)

func main() {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		return
	}

	foundPort := ""
	for _, port := range ports {
		if port.IsUSB {
			switch port.SerialNumber {
			case "USB35INCHIPSV2":
				fmt.Printf("Port: %s, (%s:%s, %s)\n", port.Name,
					port.VID, port.PID,
					port.SerialNumber)
				foundPort = port.Name
			}
		}
	}

	if foundPort == "" {
		log.Fatalf("No USB Monitor found")
	}

	dev := foundPort
	lcdInfo, err := lcd.New(dev, common.RevisionA)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer lcdInfo.Close()

	log.Printf("LCD Info: %v", lcdInfo.Revision())

	err = lcdInfo.SetBrightness(10)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	err = lcdInfo.SetOrientation(common.Landscape)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// infile, err := os.Open("example_320x480.png")
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// defer infile.Close()

	// img, _, err := image.Decode(infile)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// dc := gg.NewContextForImage(img)

	dc := gg.NewContext(lcdInfo.Width(), lcdInfo.Height())

	const S = 320
	for i := 0; i < 360; i += 15 {
		dc.Push()
		dc.SetColor(color.RGBA{255, 255, 255, 255})
		dc.RotateAbout(gg.Radians(float64(i)), S/2, S/2)
		dc.DrawEllipse(S/2, S/2, S*7/16, S/8)
		dc.Fill()
		dc.Pop()
	}
	dc.SetColor(color.RGBA{0, 0, 0, 255})
	dc.DrawString("Hello, world!", S/2, S/2)

	lcdInfo.Clear()

	for {
		x1 := rand.Float64() * float64(lcdInfo.Width())
		y1 := rand.Float64() * float64(lcdInfo.Height())
		x2 := rand.Float64() * float64(lcdInfo.Width())
		y2 := rand.Float64() * float64(lcdInfo.Height())
		r := rand.Float64()
		g := rand.Float64()
		b := rand.Float64()
		a := rand.Float64()*0.5 + 0.5
		w := rand.Float64()*4 + 1
		dc.SetRGBA(r, g, b, a)
		dc.SetLineWidth(w)
		dc.DrawLine(x1, y1, x2, y2)
		dc.Stroke()

		lcd.DisplayImage(lcdInfo, dc.Image(), 0, 0, 320, 480)
	}
}
