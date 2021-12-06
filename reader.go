package modbus

import (
	"net"
	"time"
)

type connTimedReader struct {
	conn net.Conn
}

func (to connTimedReader) TimedRead(buf []byte, toms int) (c int, err error) {
	c = 0
	dl := time.Now().Add(durationMs(toms))
	err = to.conn.SetReadDeadline(dl)
	if err != nil {
		return
	}
	c, err = to.conn.Read(buf)
	return
}
