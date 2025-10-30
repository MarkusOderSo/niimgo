package niimprint

import (
	"bytes"
	"fmt"
)

type Packet struct {
	Type byte
	Data []byte
}

func NewPacket(typ byte, data []byte) *Packet {
	return &Packet{
		Type: typ,
		Data: data,
	}
}

func (p *Packet) ToBytes() []byte {
	buf := bytes.NewBuffer(nil)

	buf.WriteByte(0x55)
	buf.WriteByte(0x55)

	buf.WriteByte(p.Type)
	buf.WriteByte(byte(len(p.Data)))

	buf.Write(p.Data)

	checksum := p.Type ^ byte(len(p.Data))
	for _, b := range p.Data {
		checksum ^= b
	}
	buf.WriteByte(checksum)

	buf.WriteByte(0xAA)
	buf.WriteByte(0xAA)

	return buf.Bytes()
}

func ParsePacket(data []byte) (*Packet, error) {
	if len(data) < 7 {
		return nil, fmt.Errorf("packet too short: %d bytes", len(data))
	}

	if data[0] != 0x55 || data[1] != 0x55 {
		return nil, fmt.Errorf("invalid header: %02x %02x", data[0], data[1])
	}

	if data[len(data)-2] != 0xAA || data[len(data)-1] != 0xAA {
		return nil, fmt.Errorf("invalid footer")
	}

	typ := data[2]
	length := int(data[3])

	if len(data) < length+7 {
		return nil, fmt.Errorf("packet size mismatch: expected %d, got %d", length+7, len(data))
	}

	checksum := typ ^ byte(length)
	for i := 4; i < 4+length; i++ {
		checksum ^= data[i]
	}

	if checksum != data[4+length] {
		return nil, fmt.Errorf("checksum mismatch: expected %02x, got %02x", data[4+length], checksum)
	}

	payload := make([]byte, length)
	copy(payload, data[4:4+length])

	return &Packet{
		Type: typ,
		Data: payload,
	}, nil
}
