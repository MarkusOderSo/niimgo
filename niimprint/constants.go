package niimprint

type InfoType byte

const (
	InfoDeviceType   InfoType = 8
	InfoSoftVersion  InfoType = 9
	InfoBattery      InfoType = 10
	InfoDeviceSerial InfoType = 11
	InfoHardVersion  InfoType = 12
)

type RequestCode byte

const (
	RequestGetInfo      RequestCode = 0x40
	RequestSetLabelType RequestCode = 0x23
	RequestSetDensity   RequestCode = 0x21
	RequestStartPrint   RequestCode = 0x01
	RequestEndPrint     RequestCode = 0xF3
	RequestStartPage    RequestCode = 0x03
	RequestEndPage      RequestCode = 0xE3
	RequestSetDimension RequestCode = 0x13
	RequestSetQuantity  RequestCode = 0x15
)
