package niimprint

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go.bug.st/serial"
)

type Transport interface {
	Read(length int) ([]byte, error)
	Write(data []byte) (int, error)
	Close() error
}

type SerialTransport struct {
	port serial.Port
}

func NewSerialTransport(portName string) (*SerialTransport, error) {
	if portName == "" || portName == "auto" {
		ports, err := serial.GetPortsList()
		if err != nil {
			return nil, fmt.Errorf("failed to list ports: %w", err)
		}

		if len(ports) == 0 {
			commonPaths := []string{"/dev/ttyACM0", "/dev/ttyACM1", "/dev/ttyUSB0", "/dev/ttyUSB1"}
			for _, path := range commonPaths {
				if _, err := os.Stat(path); err == nil {
					portName = path
					fmt.Printf("Found device at: %s\n", portName)
					break
				}
			}

			if portName == "" || portName == "auto" {
				return nil, fmt.Errorf("no serial ports detected")
			}
		} else {
			if len(ports) > 1 {
				for _, p := range ports {
					if len(p) >= 10 && p[:10] == "/dev/ttyAC" {
						portName = p
						break
					}
				}
				if portName == "" || portName == "auto" {
					msg := "multiple serial ports found, please specify one:\n"
					for _, p := range ports {
						msg += fmt.Sprintf("  - %s\n", p)
					}
					return nil, fmt.Errorf(msg)
				}
			} else {
				portName = ports[0]
			}
		}
	}

	fmt.Printf("Opening serial port: %s\n", portName)

	mode := &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port %s: %w", portName, err)
	}

	if err := port.SetDTR(true); err == nil {
		if err := port.SetRTS(true); err == nil {
			time.Sleep(100 * time.Millisecond)
		}
	}

	port.SetReadTimeout(500 * time.Millisecond)

	return &SerialTransport{port: port}, nil
}

func (t *SerialTransport) Read(length int) ([]byte, error) {
	buf := make([]byte, length)
	n, err := t.port.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (t *SerialTransport) Write(data []byte) (int, error) {
	return t.port.Write(data)
}

func (t *SerialTransport) Close() error {
	if t.port != nil {
		return t.port.Close()
	}
	return nil
}

// UsbTransport communicates with USB printer class devices (e.g. /dev/usb/lp*).
// This is required for printers like the Niimbot K3 that use the USB printer
// interface for bidirectional Niimbot protocol communication.
type UsbTransport struct {
	file *os.File
}

// NewUsbTransport opens a USB printer device. When devicePath is empty or "auto",
// it scans /dev/usb/lp0 through /dev/usb/lp7 for the first available device.
func NewUsbTransport(devicePath string) (*UsbTransport, error) {
	if devicePath == "" || devicePath == "auto" {
		for i := 0; i < 8; i++ {
			path := fmt.Sprintf("/dev/usb/lp%d", i)
			if _, err := os.Stat(path); err == nil {
				devicePath = path
				fmt.Printf("Found USB printer device at: %s\n", devicePath)
				break
			}
		}
		if devicePath == "" || devicePath == "auto" {
			return nil, fmt.Errorf("no USB printer devices found")
		}
	}

	fmt.Printf("Opening USB device: %s\n", devicePath)

	file, err := os.OpenFile(devicePath, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open USB device %s: %w", devicePath, err)
	}

	return &UsbTransport{file: file}, nil
}

// usbReadTimeout is the per-read deadline used by UsbTransport.
const usbReadTimeout = 500 * time.Millisecond

func (t *UsbTransport) Read(length int) ([]byte, error) {
	if err := t.file.SetReadDeadline(time.Now().Add(usbReadTimeout)); err != nil {
		return nil, fmt.Errorf("set read deadline: %w", err)
	}
	buf := make([]byte, length)
	n, err := t.file.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (t *UsbTransport) Write(data []byte) (int, error) {
	return t.file.Write(data)
}

func (t *UsbTransport) Close() error {
	if t.file != nil {
		return t.file.Close()
	}
	return nil
}

// NewTransport returns the appropriate transport for the given port path.
// For a path starting with "/dev/usb/lp", a UsbTransport is used (K3 and
// similar USB printer-class devices). All other paths use a SerialTransport.
// When portName is empty or "auto", USB printer devices are tried first; if
// none are found, serial port auto-detection is used (preserving the existing
// behaviour for D110/D11 printers).
func NewTransport(portName string) (Transport, error) {
	if portName != "" && portName != "auto" {
		if strings.HasPrefix(portName, "/dev/usb/lp") {
			return NewUsbTransport(portName)
		}
		return NewSerialTransport(portName)
	}

	// Auto-detect: try USB printer devices first (K3), then fall back to serial.
	if transport, err := NewUsbTransport("auto"); err == nil {
		return transport, nil
	}

	return NewSerialTransport("auto")
}
