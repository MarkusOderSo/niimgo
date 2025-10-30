# niimgo

> A Go client for Niimbot label printers (D110, D11, and compatible models)

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue.svg)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A pure Go implementation for controlling Niimbot label printers via USB serial communication. Ported from the Python library [niimprint](https://github.com/kjy00302/niimprint).

## ✨ Features

- ✅ USB serial communication (CDC ACM)
- ✅ Automatic port detection
- ✅ Support for Niimbot D110 (40x12mm labels) and D11 (15x30mm labels)
- ✅ Automatic image processing (scaling, conversion to black & white)
- ✅ Adjustable print density (1-5)
- ✅ Device information retrieval
- ✅ No external dependencies except Go standard library and serial port library

## 📦 Installation

### Download prebuilt binaries

Download the latest release for your platform from the [Releases page](https://github.com/MarkusOderSo/niimgo/releases):

- **Linux**: `niimgo-vX.X.X-linux-amd64.tar.gz` (Intel/AMD 64-bit)
- **Linux ARM**: `niimgo-vX.X.X-linux-arm64.tar.gz` (ARM 64-bit, e.g., Raspberry Pi 4)
- **Linux ARM**: `niimgo-vX.X.X-linux-arm.tar.gz` (ARM 32-bit, e.g., older Raspberry Pi)
- **macOS**: `niimgo-vX.X.X-darwin-amd64.tar.gz` (Intel Mac)
- **macOS**: `niimgo-vX.X.X-darwin-arm64.tar.gz` (Apple Silicon M1/M2/M3)
- **Windows**: `niimgo-vX.X.X-windows-amd64.zip` (Intel/AMD 64-bit)

Extract and run:
```bash
# Linux/macOS
tar xzf niimgo-*.tar.gz
sudo ./niimgo-* -port /dev/ttyACM0 image.png

# Windows
# Extract the zip and run niimgo-windows-amd64.exe
```

### Build from source

```bash
git clone https://github.com/MarkusOderSo/niimgo.git
cd niimgo
go build -o niimgo ./cmd
```

### Or install directly

```bash
go install github.com/MarkusOderSo/niimgo/cmd@latest
```

## 📋 Prerequisites

- Go 1.21 or higher
- USB connection to Niimbot printer
- Linux/macOS (Windows support via Niimbot Windows driver)

## 📐 Understanding Label Sizes

The D110 prints on **40x12mm labels**:
- **Print head width**: 12mm = 96 pixels (across the label)
- **Label length**: up to 40mm = 320 pixels (print direction)

Your image should be **96 pixels wide** and up to **320 pixels tall**.

```
┌─────────────────────────────────┐  ↑
│                                 │  │ 12mm (96 pixels wide)
│  Your image: 96 x ??? pixels   │  │
│                                 │  ↓
└─────────────────────────────────┘
←────────── 40mm (320 pixels) ──────→
        (Print direction →)
```

## 🚀 Usage

### Basic printing (D110 with 40x12mm labels)

```bash
sudo ./niimgo -port /dev/ttyACM0 examples/test_label_40x12mm_96x320.png
```

The image will be automatically scaled to 96 pixels width (matching the 12mm print head).

### With options

```bash
# Higher print density (1-5)
sudo ./niimgo -port /dev/ttyACM0 -density 5 examples/test_label_40x12mm_96x320.png

# Specify port explicitly
sudo ./niimgo -port /dev/ttyACM0 examples/test_label_40x12mm_96x320.png

# Debug mode
sudo ./niimgo -port /dev/ttyACM0 -debug examples/test_label_40x12mm_96x320.png

# Show device information only
sudo ./niimgo -port /dev/ttyACM0 -info
```

## 🎨 Image Preparation

For best results:

1. **Image width**: Create your image with 96 pixels width
2. **Height**: Up to 320 pixels (= 40mm label length)
3. **Example**: A 96x200 pixel image = 12mm wide × 25mm long

```bash
# Example: Perfectly sized image
sudo ./niimgo my_label_96x200.png

# The program automatically scales to 96 pixels width
sudo ./niimgo any_image.png
```

## 📋 Supported Labels

| Printer | Label | Print Head | Max Length | Example |
|---------|-------|------------|------------|---------|
| **D110** | 40x12mm | 96 px (12mm) | 320 px (40mm) | `sudo ./niimgo test.png` |
| **D11** | 15x30mm | 120 px (15mm) | 240 px (30mm) | `sudo ./niimgo -width 120 test.png` |

## 💡 Tips

- **Image preparation**: Use images with 96 pixels width for optimal quality
- **Auto-scaling**: The program automatically scales to 96 pixels width
- **Length**: Your image can be up to 320 pixels long (= 40mm)
- **Print density**: Higher values (4-5) for darker prints

## 🔧 Technical Details

This implementation:
- Uses **USB serial communication** via `/dev/ttyACM*` (CDC ACM)
- Follows the same protocol as the Python version
- Baud rate: 115200
- Packet format: `0x55 0x55 [TYPE] [LEN] [DATA...] [CHECKSUM] 0xAA 0xAA`
- Image format: Black/white with 1-bit per pixel (MSB = left)
- **D110**: Print head 12mm (96 pixels wide), labels up to 40mm long (320 pixels)

## 🐛 Troubleshooting

### "No serial ports detected"
```bash
# Check if device is recognized
lsusb | grep 3513

# Check available ports
ls -la /dev/ttyACM*

# Specify port explicitly
sudo ./niimgo -port /dev/ttyACM0 test_label_40x12mm_96x320.png
```

### "Permission denied"
```bash
# Run with sudo
sudo ./niimgo test_label_40x12mm_96x320.png

# OR: Add user to dialout group
sudo usermod -a -G dialout $USER
# Then log out and back in
```

### Image prints at the end of the label
- This is normal - the printer prints in the lengthwise direction
- The image appears at the end of the label as it passes through

### Image prints too small
- Make sure your image is **96 pixels wide**
- The program scales automatically, but prepared images look better
- Increase print density with `-density 5`

## 🙏 Credits

Based on the Python implementation by [kjy00302](https://github.com/kjy00302/niimprint).

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.