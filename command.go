package modbus

func (c *Command) CheckValid() error {
	switch c.Code {
	case ReadDos01, ReadDis02:
		if c.Corv < 1 || c.Corv > MaxBools {
			return formatErr("Count %d out of range [1, %d]", c.Corv, MaxBools)
		}
	case ReadWos03, ReadWis04:
		if c.Corv < 1 || c.Corv > MaxWords {
			return formatErr("Count %d out of range [1, %d]", c.Corv, MaxWords)
		}
	case WriteDos15:
		count := len(c.Bools)
		if count < 1 || count > MaxBools {
			return formatErr("Count %d out of range [1, %d]", count, MaxBools)
		}
		if uint16(count) != c.Corv {
			return formatErr("Count mismatch %d got %d", count, c.Corv)
		}
	case WriteWos16:
		count := len(c.Words)
		if count < 1 || count > MaxWords {
			return formatErr("Count %d out of range [1, %d]", count, MaxWords)
		}
		if uint16(count) != c.Corv {
			return formatErr("Count mismatch %d got %d", count, c.Corv)
		}
	case WriteDo05:
		if c.Corv != 0 && c.Corv != TrueWord {
			return formatErr("Corv invalid %04x expected %04x or %04x", c.Corv, 0, TrueWord)
		}
		return nil
	case WriteWo06:
		return nil
	default:
		return formatErr("Code unsupported %d", c.Code)
	}
	return nil
}

func (c *Command) CheckException(buf []byte) (err error) {
	_slave := buf[0]
	_code80 := buf[1]
	if _slave != c.Slave {
		err = formatErr("Slave mismatch got %02x expected %02x", _slave, c.Slave)
		return
	}
	if _code80 != (c.Code | 0x80) {
		err = formatErr("Code80 mismatch got %02x expected %02x | 0x80", _code80, c.Code)
		return
	}
	return
}

func (c *Command) CheckResponse(buf []byte) error {
	switch c.Code {
	case ReadDos01, ReadDis02, ReadWos03, ReadWis04:
		_slave := buf[0]
		_code := buf[1]
		_bytes := buf[2]
		bytes := c.ResponseBytes()
		if _slave != c.Slave {
			return formatErr("Slave mismatch got %02x expected %02x", _slave, c.Slave)
		}
		if _code != c.Code {
			return formatErr("Code mismatch got %02x expected %02x", _code, c.Code)
		}
		if _bytes != bytes {
			return formatErr("Byte count mismatch got %d expected %d", _bytes, bytes)
		}
		return nil
	case WriteDo05, WriteWo06, WriteDos15, WriteWos16:
		_slave := buf[0]
		_code := buf[1]
		_address := encodeWord(buf[2], buf[3])
		_corv := encodeWord(buf[4], buf[5])
		if _slave != c.Slave {
			return formatErr("Slave mismatch got %02x expected %02x", _slave, c.Slave)
		}
		if _code != c.Code {
			return formatErr("Code mismatch got %02x expected %02x", _code, c.Code)
		}
		if _address != c.Address {
			return formatErr("Address mismatch got %04x expected %04x", _address, c.Address)
		}
		if _corv != c.Corv {
			return formatErr("Corv mismatch got %04x expected %04x", _corv, c.Corv)
		}
		return nil
	default:
		return formatErr("Code unsupported %d", c.Code)
	}
}

func (c *Command) RequestLength() uint16 {
	switch c.Code {
	case WriteDos15:
		count := uint16(len(c.Bools))
		return 7 + uint16(bytesForBools(count))
	case WriteWos16:
		count := uint16(len(c.Words))
		return 7 + uint16(bytesForWords(count))
	default:
		return 6
	}
}

func (c *Command) ResponseLength() uint16 {
	switch c.Code {
	case ReadDos01, ReadDis02:
		return 3 + uint16(bytesForBools(c.Corv))
	case ReadWos03, ReadWis04:
		return 3 + uint16(bytesForWords(c.Corv))
	default:
		return 6
	}
}

func (c *Command) ResponseBytes() byte {
	switch c.Code {
	case ReadDos01, ReadDis02:
		return bytesForBools(c.Corv)
	case ReadWos03, ReadWis04:
		return bytesForWords(c.Corv)
	default:
		return 0
	}
}

func (c *Command) EncodeRequest(buf []byte) {
	buf[0] = c.Slave
	buf[1] = c.Code
	buf[2] = highByte(c.Address)
	buf[3] = lowByte(c.Address)
	buf[4] = highByte(c.Corv)
	buf[5] = lowByte(c.Corv)
	switch c.Code {
	case WriteDos15:
		length := bytesForBools(c.Corv)
		buf[6] = length
		encodeBools(buf[7:7+int(length)], c.Bools...)
	case WriteWos16:
		length := bytesForWords(c.Corv)
		buf[6] = length
		encodeWords(buf[7:7+int(length)], c.Words...)
	}
}

//not enough info in response packet to parse reads
func (c *Command) DecodeResponse(buf []byte, count uint16) {
	c.Slave = buf[0]
	c.Code = buf[1]
	switch c.Code {
	case ReadDos01, ReadDis02:
		c.Bools = make([]bool, count)
		decodeBools(buf[3:], c.Bools)
	case ReadWos03, ReadWis04:
		c.Words = make([]uint16, count)
		decodeWords(buf[3:], c.Words)
	case WriteDo05, WriteWo06, WriteDos15, WriteWos16:
		c.Address = encodeWord(buf[2], buf[3])
		c.Corv = encodeWord(buf[4], buf[5])
	}
}

func (c *Command) EncodeResponse(buf []byte) {
	buf[0] = c.Slave
	buf[1] = c.Code
	switch c.Code {
	case ReadDos01, ReadDis02:
		count := uint16(len(c.Bools))
		length := bytesForBools(count)
		buf[2] = length
		encodeBools(buf[3:3+int(length)], c.Bools...)
	case ReadWos03, ReadWis04:
		count := uint16(len(c.Words))
		length := bytesForWords(count)
		buf[2] = length
		encodeWords(buf[3:3+int(length)], c.Words...)
	case WriteDo05, WriteWo06, WriteDos15, WriteWos16:
		buf[2] = highByte(c.Address)
		buf[3] = lowByte(c.Address)
		buf[4] = highByte(c.Corv)
		buf[5] = lowByte(c.Corv)
	}
}

func (c *Command) DecodeRequest(buf []byte) error {
	c.Slave = buf[0]
	c.Code = buf[1]
	c.Address = encodeWord(buf[2], buf[3])
	c.Corv = encodeWord(buf[4], buf[5])
	switch c.Code {
	case WriteDos15:
		_bytes := buf[6]
		bytes := bytesForBools(c.Corv)
		if _bytes != bytes {
			return formatErr("Byte count mismatch got %d expected %d", _bytes, bytes)
		}
		c.Bools = make([]bool, c.Corv)
		decodeBools(buf[7:], c.Bools)
	case WriteWos16:
		_bytes := buf[6]
		bytes := bytesForWords(c.Corv)
		if _bytes != bytes {
			return formatErr("Byte count mismatch got %d expected %d", _bytes, bytes)
		}
		c.Words = make([]uint16, c.Corv)
		decodeWords(buf[7:], c.Words)
	}
	return nil
}
