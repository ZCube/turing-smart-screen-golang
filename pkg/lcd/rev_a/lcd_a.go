package rev_a

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/zcube/turing-smart-screen-golang/pkg/common"
	"go.bug.st/serial"
)

type Command byte

const (
	CMD_RESET           Command = 101
	CMD_CLEAR           Command = 102
	CMD_TO_BLACK        Command = 103
	CMD_SCREEN_OFF      Command = 108
	CMD_SCREEN_ON       Command = 109
	CMD_SET_BRIGHTNESS  Command = 110
	CMD_SET_ORIENTATION Command = 121
	CMD_DISPLAY_BITMAP  Command = 197
	// next generation commands
	CMD_LCD_28         Command = 40
	CMD_LCD_29         Command = 41
	CMD_HELLO          Command = 69
	CMD_SET_MIRROR     Command = 122
	CMD_DISPLAY_PIXELS Command = 195
)

func makeCommand0(cmd Command) []byte {
	byteBuffer := make([]byte, 6)
	byteBuffer[5] = byte(cmd)
	return byteBuffer
}

func makeCommand1(cmd Command, x int) []byte {
	byteBuffer := make([]byte, 6)
	byteBuffer[0] = byte(x >> 2)
	byteBuffer[1] = byte(((x & 3) << 6))
	byteBuffer[5] = byte(cmd)
	return byteBuffer
}

func makeCommand4(cmd Command, x, y, ex, ey int) []byte {
	byteBuffer := make([]byte, 6)
	byteBuffer[0] = byte(x >> 2)
	byteBuffer[1] = byte(((x & 3) << 6) + (y >> 4))
	byteBuffer[2] = byte(((y & 15) << 4) + (ex >> 6))
	byteBuffer[3] = byte(((ex & 63) << 2) + (ey >> 8))
	byteBuffer[4] = byte(ey & 255)
	byteBuffer[5] = byte(cmd)
	return byteBuffer
}

func makeHello() []byte {
	cmd := CMD_HELLO

	byteBuffer := make([]byte, 6)
	byteBuffer[0] = byte(cmd)
	byteBuffer[1] = byte(cmd)
	byteBuffer[2] = byte(cmd)
	byteBuffer[3] = byte(cmd)
	byteBuffer[4] = byte(cmd)
	byteBuffer[5] = byte(cmd)
	return byteBuffer
}

type RevisionResponse []byte

var (
	TURING_3_5     RevisionResponse = []byte{}
	USBMONITOR_3_5 RevisionResponse = []byte{0x01, 0x01, 0x01, 0x01, 0x01, 0x01}
	USBMONITOR_5   RevisionResponse = []byte{0x02, 0x02, 0x02, 0x02, 0x02, 0x02}
	USBMONITOR_7   RevisionResponse = []byte{0x03, 0x03, 0x03, 0x03, 0x03, 0x03}

	REVISION_TURING_3_5     string = "Turing 3.5"
	REVISION_USBMONITOR_3_5 string = "USB Monitor 3.5"
	REVISION_USBMONITOR_5   string = "USB Monitor 5"
	REVISION_USBMONITOR_7   string = "USB Monitor 7"
)

type LcdCommA struct {
	revision string
	width    int
	height   int
	port     serial.Port
}

func (l *LcdCommA) Width() int {
	return l.width
}

func (l *LcdCommA) Height() int {
	return l.height
}

func (l *LcdCommA) Revision() string {
	return l.revision
}

func (l *LcdCommA) Close() error {
	if l.port == nil {
		return nil
	}
	return l.port.Close()
}

func (l *LcdCommA) WriteData(byteBuffer []byte) (int, error) {
	offset := 0
	for offset < len(byteBuffer) {
		n, err := l.port.Write(byteBuffer[offset:])
		if err != nil {
			return offset + n, err
		}
		offset += n
	}
	return offset, nil
}

func (l *LcdCommA) ReadData(readSize int) ([]byte, error) {
	buffer := make([]byte, readSize)
	n, err := l.port.Read(buffer)
	if err != nil {
		return nil, err
	}
	if n != readSize {
		return nil, fmt.Errorf("Expected %v bytes, got %v", readSize, n)
	}
	return buffer, nil
}

func (l *LcdCommA) Reset() error {
	_, err := l.WriteData(makeCommand0(CMD_RESET))
	if err != nil {
		return err
	}
	err = l.Close()
	if err != nil {
		return err
	}
	return nil
}

func (l *LcdCommA) Clear() error {
	_, err := l.WriteData(makeCommand0(CMD_CLEAR))
	if err != nil {
		return err
	}
	return nil
}

func (l *LcdCommA) ScreenOff() error {
	_, err := l.WriteData(makeCommand0(CMD_SCREEN_OFF))
	if err != nil {
		return err
	}
	return nil
}

func (l *LcdCommA) ScreenOn() error {
	_, err := l.WriteData(makeCommand0(CMD_SCREEN_ON))
	if err != nil {
		return err
	}
	return nil
}

func (l *LcdCommA) SetBrightness(level int) error {
	if level < 0 || level > 100 {
		err := errors.New("Brightness level must be [0-100]")
		return err
	}

	levelAbsolute := 255 - int(float64(level)/100.0*255)
	_, err := l.WriteData(makeCommand1(CMD_SET_BRIGHTNESS, levelAbsolute))
	return err
}

func (l *LcdCommA) SetOrientation(orientation common.Orientation) error {
	width := l.Width()
	height := l.Height()

	byteBuffer := make([]byte, 11)
	byteBuffer[5] = byte(CMD_SET_ORIENTATION)
	byteBuffer[6] = byte(orientation + 100)
	byteBuffer[7] = byte(width >> 8)
	byteBuffer[8] = byte(width & 255)
	byteBuffer[9] = byte(height >> 8)
	byteBuffer[10] = byte(height & 255)
	n, err := l.WriteData(byteBuffer)
	if err != nil {
		return err
	}
	if n != 11 {
		return fmt.Errorf("Expected 11 bytes, got %v", n)
	}
	return nil
}

func (l *LcdCommA) DisplayImage(imageRGB656LE []byte, x, y, imageWidth, imageHeight int) error {
	x0, y0 := x, y
	x1, y1 := x+imageWidth-1, y+imageHeight-1

	_, err := l.WriteData(makeCommand4(CMD_DISPLAY_BITMAP, x0, y0, x1, y1))
	if err != nil {
		return err
	}

	_, err = l.WriteData(imageRGB656LE)
	if err != nil {
		return err
	}
	return nil
}

func NewLcdCommA(deviceName string) (*LcdCommA, error) {
	mode := &serial.Mode{
		BaudRate: 115200,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
		InitialStatusBits: &serial.ModemOutputBits{
			DTR: true,
			RTS: true,
		},
	}

	// https://github.com/serialport/node-serialport/issues/2243
	// mac ignore cts configure

	port, err := serial.Open(deviceName, mode)
	if err != nil {
		return nil, err
	}

	modemBits, err := port.GetModemStatusBits()
	if err != nil {
		return nil, err
	}

	bts, err := json.Marshal(modemBits)
	if err != nil {
		return nil, err
	}
	log.Printf("Modem status bits: %v", string(bts))

	if !modemBits.CTS {
		log.Println("CTS is not set : this is a Mac issue : rendering will not exactly works as expected")
		// err = errors.New("CTS is not set")
		// return nil, err
	}

	err = port.SetReadTimeout(time.Second)
	if err != nil {
		return nil, err
	}

	log.Printf("Opened port %v", deviceName)

	helloCommand := makeHello()
	_, err = port.Write(helloCommand)
	if err != nil {
		return nil, err
	}

	n := 0
	revision := make([]byte, 6)
	n, err = port.Read(revision)
	if n != 6 {
		return nil, fmt.Errorf("Expected 6 bytes, got %v", n)
	}
	if err != nil {
		return nil, err
	}

	var ret *LcdCommA

	if reflect.DeepEqual(revision, USBMONITOR_3_5) {
		ret = &LcdCommA{
			revision: REVISION_USBMONITOR_3_5,
			width:    320,
			height:   480,
			port:     port,
		}
	} else if reflect.DeepEqual(revision, USBMONITOR_5) {
		ret = &LcdCommA{
			revision: REVISION_USBMONITOR_3_5,
			width:    480,
			height:   800,
			port:     port,
		}
	} else if reflect.DeepEqual(revision, USBMONITOR_7) {
		ret = &LcdCommA{
			revision: REVISION_USBMONITOR_3_5,
			width:    600,
			height:   1024,
			port:     port,
		}
	} else {
		ret = &LcdCommA{
			revision: REVISION_USBMONITOR_3_5,
			width:    320,
			height:   480,
			port:     port,
		}
	}
	return ret, nil
}
