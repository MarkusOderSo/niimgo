# Niimbot D110 Go Client

Eine Go-Implementierung des Niimbot D110 Etikettendruckers, portiert von der Python-Bibliothek [niimprint](https://github.com/kjy00302/niimprint).

## Features

- ✅ Serielle USB-Kommunikation (wie die Original-Python-Version)
- ✅ Automatische Port-Erkennung
- ✅ Unterstützung für Niimbot D110 (40x12mm Etiketten)
- ✅ Bildverarbeitung (automatische Größenanpassung, Konvertierung zu S/W)
- ✅ Einstellbare Druckdichte (1-5)
- ✅ Geräteinformationen auslesen

## Installation

```bash
go build -o niimgo ./cmd
```

## Wichtig: Etikettengröße verstehen

Der D110 druckt auf **40x12mm Etiketten**:
- **Druckkopfbreite**: 12mm = 96 Pixel (quer über das Etikett)
- **Etikettenlänge**: bis zu 40mm = 320 Pixel (in Druckrichtung)

Ihr Bild sollte also **96 Pixel breit** und bis zu **320 Pixel lang** sein.

```
┌─────────────────────────────────┐  ↑
│                                 │  │ 12mm (96 Pixel breit)
│  Ihr Bild: 96 x ??? Pixel      │  │
│                                 │  ↓
└─────────────────────────────────┘
←────────── 40mm (320 Pixel) ──────→
        (Druckrichtung →)
```

## Verwendung

### Einfaches Drucken (D110 mit 40x12mm Etiketten)

```bash
sudo ./niimgo test.png
```

Das Bild wird automatisch auf 96 Pixel Breite skaliert (passend zum 12mm Druckkopf).

### Mit Optionen

```bash
# Höhere Druckdichte (1-5)
sudo ./niimgo -density 5 test.png

# Spezifischen Port angeben
sudo ./niimgo -port /dev/ttyACM0 test.png

# Debug-Modus
sudo ./niimgo -debug test.png

# Nur Geräteinformationen anzeigen
sudo ./niimgo -info
```

## Bildvorbereitung

Für beste Ergebnisse:

1. **Bildgröße**: Erstellen Sie Ihr Bild mit 96 Pixeln Breite
2. **Höhe**: Bis zu 320 Pixel (= 40mm Etikettenlänge)
3. **Beispiel**: Ein 96x200 Pixel Bild = 12mm breit × 25mm lang

```bash
# Beispiel: Perfekt dimensioniertes Bild
sudo ./niimgo my_label_96x200.png

# Das Programm skaliert automatisch auf 96 Pixel Breite
sudo ./niimgo any_image.png
```

## Unterstützte Etiketten

| Drucker | Etikett | Druckkopf | Max. Länge | Beispiel |
|---------|---------|-----------|------------|----------|
| **D110** | 40x12mm | 96 px (12mm) | 320 px (40mm) | `sudo ./niimgo test.png` |
| **D11** | 15x30mm | 120 px (15mm) | 240 px (30mm) | `sudo ./niimgo -width 120 test.png` |

## Tipps

- **Bild vorbereiten**: Verwenden Sie Bilder mit 96 Pixeln Breite für optimale Qualität
- **Bildgröße**: Das Programm skaliert automatisch auf 96 Pixel Breite
- **Länge**: Ihr Bild kann bis zu 320 Pixel lang sein (= 40mm)
- **Druckdichte**: Höhere Werte (4-5) für dunklere Ausdrucke

## Technische Details

Diese Implementierung:
- Verwendet **serielle USB-Kommunikation** über `/dev/ttyACM*` (CDC ACM)
- Folgt dem gleichen Protokoll wie die Python-Version
- Baudrate: 115200
- Paketformat: `0x55 0x55 [TYPE] [LEN] [DATA...] [CHECKSUM] 0xAA 0xAA`
- Bildformat: Schwarz/Weiß mit 1-Bit pro Pixel (MSB = links)
- **D110**: Druckkopf 12mm (96 Pixel breit), Etiketten bis 40mm lang (320 Pixel)

## Fehlerbehebung

### "No serial ports detected"
```bash
# Prüfen Sie, ob das Gerät erkannt wird
lsusb | grep 3513

# Prüfen Sie verfügbare Ports
ls -la /dev/ttyACM*

# Geben Sie explizit einen Port an
sudo ./niimgo -port /dev/ttyACM0 test.png
```

### "Permission denied"
```bash
# Ausführen mit sudo
sudo ./niimgo test.png

# ODER: Benutzer zur dialout-Gruppe hinzufügen
sudo usermod -a -G dialout $USER
# Danach neu anmelden
```

### Bild wird am Ende des Etiketts gedruckt
- Das ist normal - der Drucker druckt in Längsrichtung
- Das Bild erscheint am Ende des Etiketts, wie es durchläuft

### Bild wird zu klein gedruckt
- Stellen Sie sicher, dass Ihr Bild **96 Pixel breit** ist
- Das Programm skaliert automatisch, aber vorbereitete Bilder sehen besser aus
- Erhöhen Sie die Druckdichte mit `-density 5`

## Lizenz

MIT

## Credits

Basiert auf der Python-Implementierung von [kjy00302](https://github.com/kjy00302/niimprint)
