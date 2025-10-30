package niimprint

import (
	"fmt"
	"os"
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
