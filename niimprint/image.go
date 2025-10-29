package niimprint

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"math"
	"time"
)

// PrintImage prints an image to the printer
func (c *PrinterClient) PrintImage(img image.Image, density int) error {
	if err := c.SetLabelDensity(density); err != nil {
		return fmt.Errorf("set density: %w", err)
	}

	if err := c.SetLabelType(1); err != nil {
		return fmt.Errorf("set label type: %w", err)
	}

	if err := c.StartPrint(); err != nil {
		return fmt.Errorf("start print: %w", err)
	}

	if err := c.StartPagePrint(); err != nil {
		return fmt.Errorf("start page: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if err := c.SetDimension(height, width); err != nil {
		return fmt.Errorf("set dimension: %w", err)
	}

	// Encode and send image
	packets := encodeImage(img)
	fmt.Printf("Sending %d image rows...\n", len(packets))

	for i, pkt := range packets {
		if err := c.SendImageData(pkt); err != nil {
			return fmt.Errorf("send image data row %d: %w", i, err)
		}

		// Small delay between rows
		time.Sleep(5 * time.Millisecond)

		if (i+1)%10 == 0 {
			fmt.Printf("Sent %d/%d rows\n", i+1, len(packets))
		}
	}

	fmt.Println("Image data sent, finalizing...")

	if err := c.EndPagePrint(); err != nil {
		return fmt.Errorf("end page: %w", err)
	}

	time.Sleep(300 * time.Millisecond)

	// Try to end print, retry if needed
	fmt.Println("Ending print job...")
	for i := 0; i < 20; i++ {
		if success, err := c.EndPrint(); err == nil && success {
			fmt.Println("Print job completed!")
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("failed to complete print job (printer may still be processing)")
}

// encodeImage converts image to Niimbot packet format
func encodeImage(img image.Image) []*Packet {
	// Convert to grayscale, invert, then convert to black/white
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create black and white image
	bwImg := image.NewGray(bounds)

	// Convert to grayscale and invert
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			oldColor := img.At(x, y)
			grayColor := color.GrayModel.Convert(oldColor).(color.Gray)
			// Invert: 255 - value (white becomes black, black becomes white)
			invertedValue := 255 - grayColor.Y
			bwImg.SetGray(x, y, color.Gray{Y: invertedValue})
		}
	}

	// Apply threshold to convert to pure black and white
	threshold := uint8(128)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			gray := bwImg.GrayAt(x, y).Y
			if gray > threshold {
				bwImg.SetGray(x, y, color.Gray{Y: 255}) // White
			} else {
				bwImg.SetGray(x, y, color.Gray{Y: 0}) // Black
			}
		}
	}

	packets := make([]*Packet, 0, height)

	// Encode each line
	for y := 0; y < height; y++ {
		// Convert line to bits
		bytesPerLine := int(math.Ceil(float64(width) / 8.0))
		lineData := make([]byte, bytesPerLine)

		for x := 0; x < width; x++ {
			pixel := bwImg.GrayAt(x, y).Y
			// Bit = 1 for white pixel, 0 for black pixel
			if pixel > 0 {
				byteIdx := x / 8
				bitIdx := 7 - (x % 8)
				lineData[byteIdx] |= (1 << bitIdx)
			}
		}

		// Create packet header: row number (2 bytes), counts (3 bytes), flag (1 byte)
		header := new(bytes.Buffer)
		binary.Write(header, binary.BigEndian, uint16(y))
		header.Write([]byte{0, 0, 0}) // counts - always 0
		header.WriteByte(1)           // flag

		// Combine header and line data
		packetData := append(header.Bytes(), lineData...)

		packets = append(packets, NewPacket(0x85, packetData))
	}

	return packets
}
