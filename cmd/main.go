package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

	"niimgo/niimprint"

	"github.com/nfnt/resize"
)

func main() {
	portFlag := flag.String("port", "auto", "Serial port (auto for auto-detection)")
	densityFlag := flag.Int("density", 3, "Print density (1-5, default 3)")
	widthFlag := flag.Int("width", 96, "Print head width in pixels (96 for D110 = 12mm)")
	maxHeightFlag := flag.Int("maxheight", 320, "Maximum print length in pixels (320 = 40mm)")
	debugFlag := flag.Bool("debug", false, "Enable debug output")
	infoFlag := flag.Bool("info", false, "Show device info only")

	flag.Parse()

	if flag.NArg() < 1 && !*infoFlag {
		fmt.Println("Niimbot D110 Go Client")
		fmt.Println("======================")
		fmt.Println("\nUsage: niimgo [options] <image_file>")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  niimgo test.png                    # For 40x12mm labels")
		fmt.Println("  niimgo -density 5 test.png")
		fmt.Println("  niimgo -info")
		fmt.Println("\nNote: D110 prints 12mm wide (96 pixels), up to 40mm long (320 pixels)")
		fmt.Println("      Your image should be 96 pixels wide (or will be scaled)")
		os.Exit(1)
	}

	// Validate density
	if *densityFlag < 1 || *densityFlag > 5 {
		log.Fatalf("Density must be between 1 and 5")
	}

	// Create transport
	log.Println("Connecting to Niimbot printer...")
	transport, err := niimprint.NewSerialTransport(*portFlag)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer transport.Close()

	// Create client
	client := niimprint.NewPrinterClient(transport)
	client.SetDebug(*debugFlag)
	defer client.Close()

	log.Println("Connected successfully!")

	// Get device info
	if *infoFlag || *debugFlag {
		log.Println("\nDevice Information:")
		log.Println("-------------------")

		if info, err := client.GetInfo(niimprint.InfoDeviceSerial); err == nil {
			log.Printf("Serial Number: %v", info)
		} else if *debugFlag {
			log.Printf("Serial Number: error - %v", err)
		}

		if info, err := client.GetInfo(niimprint.InfoDeviceType); err == nil {
			log.Printf("Device Type: %v", info)
		} else if *debugFlag {
			log.Printf("Device Type: error - %v", err)
		}

		if info, err := client.GetInfo(niimprint.InfoSoftVersion); err == nil {
			log.Printf("Software Version: %.2f", info)
		} else if *debugFlag {
			log.Printf("Software Version: error - %v", err)
		}

		if info, err := client.GetInfo(niimprint.InfoHardVersion); err == nil {
			log.Printf("Hardware Version: %.2f", info)
		} else if *debugFlag {
			log.Printf("Hardware Version: error - %v", err)
		}

		if info, err := client.GetInfo(niimprint.InfoBattery); err == nil {
			log.Printf("Battery Level: %v%%", info)
		} else if *debugFlag {
			log.Printf("Battery Level: error - %v", err)
		}

		log.Println()
	}

	if *infoFlag {
		return
	}

	// Load and print image
	imagePath := flag.Arg(0)

	log.Printf("Loading image: %s", imagePath)
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatalf("Failed to open image: %v", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		log.Fatalf("Failed to decode image: %v", err)
	}

	log.Printf("Image format: %s", format)

	// Get original image dimensions
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	log.Printf("Original image size: %dx%d pixels", imgWidth, imgHeight)
	log.Printf("Label size: 40x12mm (print head width: 12mm = 96 pixels)")

	// Scale to fit print head width (96 pixels = 12mm)
	targetWidth := *widthFlag
	var finalImg image.Image = img

	if imgWidth != targetWidth {
		// Calculate new height keeping aspect ratio
		aspectRatio := float64(imgHeight) / float64(imgWidth)
		targetHeight := uint(float64(targetWidth) * aspectRatio)

		// Limit to max label length
		maxHeight := uint(*maxHeightFlag)
		if targetHeight > maxHeight {
			log.Printf("Warning: Image length (%d pixels) exceeds label length (%d pixels), will be cropped", targetHeight, maxHeight)
			targetHeight = maxHeight
		}

		log.Printf("Scaling image to %dx%d pixels (width=12mm, length=%dmm)", targetWidth, targetHeight, int(float64(targetHeight)/8.0))
		finalImg = resize.Resize(uint(targetWidth), targetHeight, img, resize.Lanczos3)
	} else {
		finalHeight := imgHeight
		if finalHeight > *maxHeightFlag {
			log.Printf("Warning: Image length (%d pixels) exceeds label length (%d pixels)", finalHeight, *maxHeightFlag)
			finalImg = resize.Resize(uint(targetWidth), uint(*maxHeightFlag), img, resize.Lanczos3)
		}
	}

	finalBounds := finalImg.Bounds()
	log.Printf("Final print size: %dx%d pixels (%.1fx%.1fmm)",
		finalBounds.Dx(), finalBounds.Dy(),
		float64(finalBounds.Dx())/8.0, float64(finalBounds.Dy())/8.0)

	// Print
	log.Printf("Starting print with density %d...", *densityFlag)
	if err := client.PrintImage(finalImg, *densityFlag); err != nil {
		log.Fatalf("Print failed: %v", err)
	}

	log.Println("✓ Print completed successfully!")
}

// rotateImage90 rotates an image 90 degrees clockwise
func rotateImage90(src image.Image) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create new image with swapped dimensions
	dst := image.NewRGBA(image.Rect(0, 0, height, width))

	// Rotate 90° clockwise: (x, y) -> (height - 1 - y, x)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dst.Set(height-1-y, x, src.At(x, y))
		}
	}

	return dst
}
