package niimprint

// InfoType represents different device information types
type InfoType byte

const (
	InfoDensity      InfoType = 1
	InfoPrintSpeed   InfoType = 2
	InfoLabelType    InfoType = 3
	InfoLanguageType InfoType = 6
	InfoAutoShutdown InfoType = 7
	InfoDeviceType   InfoType = 8
	InfoSoftVersion  InfoType = 9
	InfoBattery      InfoType = 10
	InfoDeviceSerial InfoType = 11
	InfoHardVersion  InfoType = 12
)

// RequestCode represents different request command codes
type RequestCode byte

const (
	RequestGetInfo        RequestCode = 0x40 // 64
	RequestGetRFID        RequestCode = 0x1A // 26
	RequestHeartbeat      RequestCode = 0xDC // 220
	RequestSetLabelType   RequestCode = 0x23 // 35
	RequestSetDensity     RequestCode = 0x21 // 33
	RequestStartPrint     RequestCode = 0x01 // 1
	RequestEndPrint       RequestCode = 0xF3 // 243
	RequestStartPage      RequestCode = 0x03 // 3
	RequestEndPage        RequestCode = 0xE3 // 227
	RequestAllowPrint     RequestCode = 0x20 // 32
	RequestSetDimension   RequestCode = 0x13 // 19
	RequestSetQuantity    RequestCode = 0x15 // 21
	RequestGetPrintStatus RequestCode = 0xA3 // 163
)

// ResponseOffset calculates the response code for a request
func (r RequestCode) ResponseCode(offset byte) byte {
	return byte(r) + offset
}
