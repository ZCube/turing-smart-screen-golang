package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"log"

	_ "image/jpeg"
	_ "image/png"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
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

	err = lcdInfo.SetBrightness(100)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	err = lcdInfo.SetOrientation(common.Landscape)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// create context
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.Flag("window-size", fmt.Sprintf("%d,%d", lcdInfo.Width(), lcdInfo.Height())),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var buf []byte
	lcdInfo.Clear()

	if err := chromedp.Run(ctx, loadPage(`https://www.google.com/`)); err != nil {
		log.Fatal(err)
	}

	for {
		log.Println("Capturing page")
		if err := chromedp.Run(ctx, capturePage(&buf)); err != nil {
			log.Fatal(err)
		}

		img, _, err := image.Decode(bytes.NewReader(buf))
		if err != nil {
			log.Fatalln(err)
		}

		lcd.DisplayImage(lcdInfo, img, 0, 0, lcdInfo.Width(), lcdInfo.Height())
	}
}

func loadPage(urlstr string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
	}
}

func capturePage(res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, err := page.CaptureScreenshot().
				WithFormat(page.CaptureScreenshotFormatPng).
				WithCaptureBeyondViewport(true).
				WithFromSurface(true).
				Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			return nil
		}),
	}
}
