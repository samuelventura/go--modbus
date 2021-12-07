package modbus

type tcpProtocol struct {
	tid uint16
}

func (p *tcpProtocol) Finally() {
	p.tid++
}

func (p *tcpProtocol) ExceptionLen() int {
	return 9
}

func (p *tcpProtocol) MakeBuffers(length uint16) (freq []byte, req []byte) {
	size := int(length)
	freq = make([]byte, size+6)
	req = freq[6:]
	return
}

func (p *tcpProtocol) WrapBuffer(buf []byte, length uint16) {
	buf[0] = highByte(p.tid)
	buf[1] = lowByte(p.tid)
	buf[2] = 0
	buf[3] = 0
	buf[4] = highByte(length)
	buf[5] = lowByte(length)
}

func (p *tcpProtocol) CheckWrapper(buf []byte, length uint16) error {
	_tid := encodeWord(buf[0], buf[1])
	_proto := encodeWord(buf[2], buf[3])
	_length := encodeWord(buf[4], buf[5])
	if _tid != p.tid {
		return formatErr("tid mismatch got %04x expected %04x", _tid, p.tid)
	}
	if _proto != 0 {
		return formatErr("proto mismatch got %04x expected %04x", _proto, 0)
	}
	if _length != length {
		return formatErr("length mismatch got %d expected %d", _length, length)
	}
	return nil
}

func (p *tcpProtocol) Scan(t Transport) (c *Command, err error) {
	head := make([]byte, 6)
	c1, err := t.TimedRead(head, -1)
	if err != nil {
		return
	}
	if c1 < 6 {
		err = formatErr("partial head %d of %d", c1, 6)
		return
	}
	p.tid = encodeWord(head[0], head[1])
	_proto := encodeWord(head[2], head[3])
	_length := encodeWord(head[4], head[5])
	if _proto != 0 {
		err = formatErr("proto mismatch got %d expected %d", _proto, 0)
		return
	}
	if _length < 6 { //request reads are 6 and writes are >=6
		err = formatErr("length mismatch got %d expected >=%d", _length, 6)
		return
	}
	//should come in single packet
	pending := int(_length)
	buf := make([]byte, pending)
	c2, err := t.TimedRead(buf, 0)
	if err != nil {
		return
	}
	if c2 < pending {
		err = formatErr("partial body %d of %d", c1+c2, pending+6)
		return
	}
	code := buf[1]
	corv := encodeWord(buf[4], buf[5]) //count or value
	length := requestLength(code, corv)
	if _length != length { //reads are 6 and writes are >=6
		err = formatErr("length mismatch got %d expected %d", _length, length)
		return
	}
	c = &Command{}
	err = c.DecodeRequest(buf)
	return
}
