package modbus

func requestLength(code byte, count uint16) uint16 {
	switch code {
	case WriteDos15:
		return 7 + uint16(bytesForBools(count))
	case WriteWos16:
		return 7 + uint16(bytesForWords(count))
	default:
		return 6
	}
}

func encodeBools(buf []byte, values ...bool) {
	for i := range buf {
		buf[i] = 0
	}
	for i, v := range values {
		if v {
			bi := i / 8
			buf[bi] |= byte(1 << (i % 8))
		}
	}
}

func decodeBools(buf []byte, bools []bool) {
	c := len(bools)
	for i, b := range buf {
		e := 8*i + 8
		if e > c {
			e = c
		}
		byteToBools(b, bools[8*i:e])
	}
}

func decodeWords(buf []byte, words []uint16) {
	for i := range words {
		words[i] = encodeWord(buf[2*i+0], buf[2*i+1])
	}
}

func byteToBools(b byte, bools []bool) {
	for i := range bools {
		bools[i] = ((b >> i) & 0x01) == 1
	}
}

func encodeWords(buf []byte, values ...uint16) {
	for i := range buf {
		buf[i] = 0
	}
	for i, v := range values {
		buf[2*i+0] = highByte(v)
		buf[2*i+1] = lowByte(v)
	}
}

func bytesForBools(count uint16) byte {
	return byte((count-1)/8 + 1)
}

func bytesForWords(count uint16) byte {
	return byte(2 * count)
}

func highByte(word uint16) byte {
	return byte((word >> 8) & 0xff)
}

func lowByte(word uint16) byte {
	return byte((word >> 0) & 0xff)
}

func encodeWord(high byte, low byte) uint16 {
	return uint16(((uint16(high) << 8) & 0xFF00) | (uint16(low) & 0xff))
}
