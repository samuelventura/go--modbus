package modbus

import (
	"io"
)

type ioTransport struct {
	closer  io.Closer
	writer  io.Writer
	reader  TimedReader
	discard bool
}

func (t *ioTransport) Close() (err error) {
	err = t.closer.Close()
	return
}

func (t *ioTransport) DiscardOn() {
	t.discard = true
}

func (t *ioTransport) DiscardIf() (err error) {
	if !t.discard {
		return
	}
	t.discard = false
	buf := make([]byte, 256)
	c, err := t.reader.TimedRead(buf)
	for c > 0 && err != io.EOF {
		c, err = t.reader.TimedRead(buf)
	}
	//only report EOF
	if err != io.EOF {
		err = nil
	}
	return
}

func (t *ioTransport) TimedRead(buf []byte, toms int) (count int, err error) {
	toms64 := int64(toms)
	start := unixMillis()
	total := len(buf)
	readc := 0
	for count < total {
		readc, err = t.reader.TimedRead(buf[count:])
		if readc > 0 {
			count += readc
		} else {
			if count > 0 {
				if err == nil {
					//break at middle of packet detected
					err = formatErr("read inter timeout %d of %d", count, total)
				}
				return
			}
		}
		if err == io.EOF {
			return
		}
		//keep reading, ignore timeout if readc > 0
		if count < total && toms >= 0 && readc <= 0 {
			now := unixMillis()
			if now-start >= toms64 {
				err = formatErr("read total timeout %d of %d", count, total)
				return
			}
		}
	}
	return
}

func (t *ioTransport) Write(buf []byte) (c int, err error) {
	c, err = t.writer.Write(buf)
	return
}
