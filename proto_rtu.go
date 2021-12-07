package modbus

import (
	"bytes"
)

type rtuProtocol struct {
}

func (p *rtuProtocol) Finally() {
}

func (p *rtuProtocol) ExceptionLen() int {
	return 5
}

func (p *rtuProtocol) MakeBuffers(length uint16) (freq []byte, req []byte) {
	size := int(length)
	freq = make([]byte, size+2)
	req = freq[:size]
	return
}

//CRC is little endian
//http://modbus.org/docs/Modbus_over_serial_line_V1_02.pdf page 13
func (p *rtuProtocol) WrapBuffer(buf []byte, length uint16) {
	offset := int(length)
	crc := crc16(buf[0:offset])
	buf[offset+0] = lowByte(crc)
	buf[offset+1] = highByte(crc)
}

func (p *rtuProtocol) CheckWrapper(buf []byte, length uint16) error {
	offset := int(length)
	crc := crc16(buf[0:offset])
	_crc := encodeWord(buf[offset+1], buf[offset+0])
	if _crc != crc {
		return formatErr("crc mismatch got %04x expected %04x", _crc, crc)
	}
	return nil
}

func (p *rtuProtocol) Scan(t Transport) (c *Command, err error) {
	head := make([]byte, 6)
	c1, err := t.TimedRead(head, -1)
	if err != nil {
		return
	}
	if c1 < 6 {
		err = formatErr("partial head %d of %d", c1, 6)
		return
	}
	code := head[1]
	corv := encodeWord(head[4], head[5]) //count or value
	length := int(requestLength(code, corv))
	//should come in single packet
	pending := length - 4 // +2 crc
	fbuf := head
	if pending > 0 {
		c2 := 0
		buf := make([]byte, pending)
		c2, err = t.TimedRead(buf, 0)
		if err != nil {
			return
		}
		if c2 < pending {
			err = formatErr("partial scan %d of %d", c1+c2, length)
			return
		}
		fbuf = bytes.Join([][]byte{head, buf}, nil)
	}
	length = len(fbuf)
	_crc := encodeWord(fbuf[length-1], fbuf[length-2])
	buf := fbuf[:length-2]
	crc := crc16(buf)
	if _crc != crc {
		err = formatErr("crc mismatch got %04x expected %04x", _crc, crc)
		return
	}
	c = &Command{}
	err = c.DecodeRequest(buf)
	return
}

func crc16(bytes []byte) (crc uint16) {
	crc = 0xFFFF
	for _, b := range bytes {
		crc ^= uint16(b)
		for i := 8; i > 0; i-- {
			if (crc & 0x0001) != 0 {
				crc >>= 1
				crc ^= 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return
}
