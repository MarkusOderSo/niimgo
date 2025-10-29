package niimprint

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"time"
)

// PrinterClient handles communication with Niimbot printer
type PrinterClient struct {
	transport Transport
	buffer    []byte
	debug     bool
}

// NewPrinterClient creates a new printer client
func NewPrinterClient(transport Transport) *PrinterClient {
	return &PrinterClient{
		transport: transport,
		buffer:    make([]byte, 0),
		debug:     false,
	}
}

// SetDebug enables or disables debug logging
func (c *PrinterClient) SetDebug(debug bool) {
	c.debug = debug
}

// Close closes the transport
func (c *PrinterClient) Close() error {
	return c.transport.Close()
}

// send sends a packet without waiting for response
func (c *PrinterClient) send(packet *Packet) error {
	data := packet.ToBytes()
	if c.debug {
		c.logBuffer("send", data)
	}
	_, err := c.transport.Write(data)
	return err
}

// recv receives and parses packets from the buffer
func (c *PrinterClient) recv() ([]*Packet, error) {
	// Read data from transport (with timeout)
	data, err := c.transport.Read(1024)
	if err != nil {
		// Timeout is not a fatal error for recv
		if len(c.buffer) > 0 {
			// Try to parse what we have
			return c.parseBufferedPackets(), nil
		}
		return nil, err
	}

	// Append to buffer
	c.buffer = append(c.buffer, data...)

	// Parse packets from buffer
	return c.parseBufferedPackets(), nil
}

// parseBufferedPackets parses packets from the internal buffer
func (c *PrinterClient) parseBufferedPackets() []*Packet {
	packets := make([]*Packet, 0)

	for len(c.buffer) > 4 {
		// Need at least 7 bytes for a valid packet
		if len(c.buffer) < 7 {
			break
		}

		// Check if we have a valid header
		if c.buffer[0] != 0x55 || c.buffer[1] != 0x55 {
			// Skip invalid byte
			c.buffer = c.buffer[1:]
			continue
		}

		// Get packet length
		pktLen := int(c.buffer[3]) + 7 // data length + header(2) + type(1) + len(1) + checksum(1) + footer(2)

		// Check if we have enough data
		if len(c.buffer) < pktLen {
			break
		}

		// Parse packet
		packet, err := ParsePacket(c.buffer[:pktLen])
		if err != nil {
			// Skip this byte and try again
			c.buffer = c.buffer[1:]
			continue
		}

		if c.debug {
			c.logBuffer("recv", c.buffer[:pktLen])
		}

		packets = append(packets, packet)
		c.buffer = c.buffer[pktLen:]
	}

	return packets
}

// transceive sends a packet and waits for matching response
func (c *PrinterClient) transceive(reqCode RequestCode, data []byte, respOffset byte, retries int, waitTime time.Duration) (*Packet, error) {
	respCode := byte(reqCode) + respOffset

	packet := NewPacket(byte(reqCode), data)
	if err := c.send(packet); err != nil {
		return nil, fmt.Errorf("failed to send: %w", err)
	}

	// Clear buffer before waiting for response
	c.buffer = c.buffer[:0]

	// Wait for response with retries
	for i := 0; i < retries; i++ {
		time.Sleep(waitTime)

		packets, err := c.recv()
		if err != nil && len(packets) == 0 {
			// No data yet, continue
			continue
		}

		for _, pkt := range packets {
			if pkt.Type == 219 {
				return nil, fmt.Errorf("received error packet (type 219)")
			}
			if pkt.Type == 0 {
				return nil, fmt.Errorf("received unsupported packet type 0")
			}
			if pkt.Type == respCode {
				return pkt, nil
			}
			// Log unexpected packet
			if c.debug {
				log.Printf("Unexpected packet type: 0x%02x (waiting for 0x%02x)", pkt.Type, respCode)
			}
		}
	}

	return nil, fmt.Errorf("timeout: no response for request code 0x%02x (expected response 0x%02x) after %d retries", reqCode, respCode, retries)
}

// logBuffer logs a buffer in hex format
func (c *PrinterClient) logBuffer(prefix string, data []byte) {
	msg := ""
	for i, b := range data {
		if i > 0 {
			msg += ":"
		}
		msg += fmt.Sprintf("%02x", b)
	}
	log.Printf("%s: %s", prefix, msg)
}

// GetInfo retrieves device information
func (c *PrinterClient) GetInfo(infoType InfoType) (interface{}, error) {
	packet, err := c.transceive(RequestGetInfo, []byte{byte(infoType)}, byte(infoType), 10, 100*time.Millisecond)
	if err != nil {
		return nil, err
	}

	if len(packet.Data) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	switch infoType {
	case InfoDeviceSerial:
		return fmt.Sprintf("%x", packet.Data), nil
	case InfoSoftVersion, InfoHardVersion:
		val := packetToInt(packet)
		return float64(val) / 100.0, nil
	default:
		return packetToInt(packet), nil
	}
}

// packetToInt converts packet data to integer (big endian)
func packetToInt(p *Packet) int {
	if len(p.Data) == 0 {
		return 0
	}
	var val int
	for _, b := range p.Data {
		val = (val << 8) | int(b)
	}
	return val
}

// SetLabelType sets the label type (1=gap, 2=continuous, 3=black mark)
func (c *PrinterClient) SetLabelType(labelType int) error {
	if labelType < 1 || labelType > 3 {
		return fmt.Errorf("label type must be between 1 and 3")
	}

	packet, err := c.transceive(RequestSetLabelType, []byte{byte(labelType)}, 16, 10, 100*time.Millisecond)
	if err != nil {
		return err
	}

	if len(packet.Data) == 0 || packet.Data[0] == 0 {
		return fmt.Errorf("failed to set label type")
	}

	return nil
}

// SetLabelDensity sets print density (1-5, where 5 is darkest)
func (c *PrinterClient) SetLabelDensity(density int) error {
	if density < 1 || density > 5 {
		return fmt.Errorf("density must be between 1 and 5")
	}

	packet, err := c.transceive(RequestSetDensity, []byte{byte(density)}, 16, 10, 100*time.Millisecond)
	if err != nil {
		return err
	}

	if len(packet.Data) == 0 || packet.Data[0] == 0 {
		return fmt.Errorf("failed to set density")
	}

	return nil
}

// StartPrint starts a print job
func (c *PrinterClient) StartPrint() error {
	packet, err := c.transceive(RequestStartPrint, []byte{0x01}, 1, 10, 100*time.Millisecond)
	if err != nil {
		return err
	}

	if len(packet.Data) == 0 || packet.Data[0] == 0 {
		return fmt.Errorf("failed to start print")
	}

	return nil
}

// EndPrint ends a print job
func (c *PrinterClient) EndPrint() (bool, error) {
	packet, err := c.transceive(RequestEndPrint, []byte{0x01}, 1, 10, 100*time.Millisecond)
	if err != nil {
		return false, err
	}

	return len(packet.Data) > 0 && packet.Data[0] != 0, nil
}

// StartPagePrint starts a page
func (c *PrinterClient) StartPagePrint() error {
	packet, err := c.transceive(RequestStartPage, []byte{0x01}, 1, 10, 100*time.Millisecond)
	if err != nil {
		return err
	}

	if len(packet.Data) == 0 || packet.Data[0] == 0 {
		return fmt.Errorf("failed to start page")
	}

	return nil
}

// EndPagePrint ends a page
func (c *PrinterClient) EndPagePrint() error {
	packet, err := c.transceive(RequestEndPage, []byte{0x01}, 1, 10, 100*time.Millisecond)
	if err != nil {
		return err
	}

	if len(packet.Data) == 0 || packet.Data[0] == 0 {
		return fmt.Errorf("failed to end page")
	}

	return nil
}

// SetDimension sets print dimensions (width, height in pixels)
func (c *PrinterClient) SetDimension(width, height int) error {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint16(width))
	binary.Write(buf, binary.BigEndian, uint16(height))

	packet, err := c.transceive(RequestSetDimension, buf.Bytes(), 1, 10, 100*time.Millisecond)
	if err != nil {
		return err
	}

	if len(packet.Data) == 0 || packet.Data[0] == 0 {
		return fmt.Errorf("failed to set dimension")
	}

	return nil
}

// SetQuantity sets number of copies
func (c *PrinterClient) SetQuantity(quantity int) error {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint16(quantity))

	packet, err := c.transceive(RequestSetQuantity, buf.Bytes(), 1, 10, 100*time.Millisecond)
	if err != nil {
		return err
	}

	if len(packet.Data) == 0 || packet.Data[0] == 0 {
		return fmt.Errorf("failed to set quantity")
	}

	return nil
}

// SendImageData sends image data packet
func (c *PrinterClient) SendImageData(packet *Packet) error {
	if err := c.send(packet); err != nil {
		return err
	}

	// For image data, we don't always wait for response
	// The Python implementation just sends without checking response
	return nil
}
