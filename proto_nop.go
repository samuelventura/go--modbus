package modbus

import (
	"bytes"
)

type nopProtocol struct {
}

func (p *nopProtocol) Finally() {
}

func (p *nopProtocol) ExceptionLen() int {
	return 3
}

func (p *nopProtocol) MakeBuffers(length uint16) (freq []byte, req []byte) {
	freq = make([]byte, length)
	req = freq
	return
}

func (p *nopProtocol) WrapBuffer(buf []byte, length uint16) {
}

func (p *nopProtocol) CheckWrapper(buf []byte, length uint16) error {
	return nil
}

func (p *nopProtocol) Scan(t Transport, qtms int) (c *Command, err error) {
	head := make([]byte, 6)
	c1, err := t.TimedRead(head, -1, qtms)
	if err != nil {
		return
	}
	if c1 < 6 {
		err = formatErr("Partial head %d of %d", c1, 6)
		return
	}
	code := head[1]
	corv := encodeWord(head[4], head[5]) //count or value
	length := int(requestLength(code, corv))
	pending := length - 6
	fbuf := head
	if pending > 0 {
		c2 := 0
		//should come in single packet
		buf := make([]byte, pending)
		c2, err = t.TimedRead(buf, 0, qtms)
		if err != nil {
			return
		}
		if c2 < pending {
			err = formatErr("Partial scan %d of %d", c1+c2, length)
			return
		}
		fbuf = bytes.Join([][]byte{head, buf}, nil)
	}
	c = &Command{}
	err = c.DecodeRequest(fbuf)
	return
}
