# turing-smart-screen-golang
Turing Smart Screen library for Golang.

## Usage
With gg (https://github.com/fogleman/gg)
```go
import (
	"github.com/zcube/turing-smart-screen-golang/pkg/common"
	"github.com/zcube/turing-smart-screen-golang/pkg/lcd"
)


// open the device
lcdInfo, err := lcd.New(dev, common.RevisionA)
if err != nil {
    log.Fatalf("Error: %v", err)
}
defer lcdInfo.Close()

// create a new context for drawing
dc := gg.NewContext(lcdInfo.Width(), lcdInfo.Height())

// clear the screen
lcdInfo.Clear()

for {
    // draw a random line
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

    // display the image
    lcd.DisplayImage(lcdInfo, dc.Image(), 0, 0, 320, 480)
}
```

## Example
```bash
go run examples/simple/main.go
```

## Warning
* MacOS is not supported. There are some problems with serial communication. Maybe it's a problem with the driver.
