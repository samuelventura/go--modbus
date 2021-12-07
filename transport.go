package modbus

import (
	"io"
)

type ioTransport struct {
	closer io.Closer
	writer io.Writer
	reader TimedReader
	eflag  bool
}

func (t *ioTransport) Close() (err error) {
	err = t.closer.Close()
	return
}

func (t *ioTransport) SetError(errf bool) {
	t.eflag = errf
}

func (t *ioTransport) Discard(qtms int) (err error) {
	if !t.eflag {
		return
	}
	t.eflag = false
	buf := make([]byte, 1)
	c, err := t.reader.TimedRead(buf, qtms)
	for c > 0 && err != io.EOF {
		c = 0
		c, err = t.reader.TimedRead(buf, qtms)
	}
	//only report EOF
	if err != io.EOF {
		err = nil
	}
	return
}

func (t *ioTransport) TimedRead(buf []byte, toms int, qtms int) (count int, err error) {
	if qtms <= 0 {
		err = formatErr("quiet timeout must be positive %d", qtms)
		return
	}
	toms64 := int64(toms)
	start := unixMillis()
	total := len(buf)
	count = 0
	for count < total {
		c := 0
		c, err = t.reader.TimedRead(buf[count:], qtms)
		if c > 0 {
			count += c
		} else {
			if count > 0 {
				if err == nil {
					err = formatErr("read inter timeout %d of %d", count, total)
				}
				return
			}
		}
		if err == io.EOF {
			return
		}
		if count < total && c <= 0 && toms >= 0 {
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
