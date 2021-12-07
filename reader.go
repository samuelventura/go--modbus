package modbus

import (
	"net"
	"time"
)

type connTimedReader struct {
	conn net.Conn
}

func (to connTimedReader) TimedRead(buf []byte) (count int, err error) {
	dl := time.Now().Add(durationMs(ReadToMs))
	err = to.conn.SetReadDeadline(dl)
	if err != nil {
		return
	}
	readc, err := to.conn.Read(buf)
	//do not return negatives
	if readc > 0 {
		count += readc
	}
	return
}
