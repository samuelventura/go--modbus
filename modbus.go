package modbus

import (
	"io"
	"net"
	"time"
)

func EnableTrace(enable bool) {
	traceEnabled = enable
}

func NewNopProtocol() Protocol {
	return &nopProtocol{}
}

func NewTcpProtocol() Protocol {
	return &tcpProtocol{}
}

func NewRtuProtocol() Protocol {
	return &rtuProtocol{}
}

func NewMaster(proto Protocol, trans Transport, toms int) CloseableMaster {
	exec := NewTransportExecutor(proto, trans, toms)
	return NewCloseableMaster(exec, trans)
}

func NewRtuMaster(trans Transport, toms int) CloseableMaster {
	return NewMaster(&rtuProtocol{}, trans, toms)
}

func NewTcpMaster(trans Transport, toms int) CloseableMaster {
	return NewMaster(&tcpProtocol{}, trans, toms)
}

func NewTcpTransport(address string, toms int) (trans Transport, err error) {
	to := time.Duration(toms) * time.Millisecond
	conn, err := net.DialTimeout("tcp", address, to)
	if err != nil {
		return
	}
	trans = NewConnTransport(conn)
	return
}

func NewConnTimedReader(conn net.Conn) TimedReader {
	return &connTimedReader{conn}
}

func NewConnTransport(conn net.Conn) Transport {
	reader := NewConnTimedReader(conn)
	return NewIoTransport(reader, conn)
}

func NewIoTransport(reader TimedReader, writerCloser io.WriteCloser) Transport {
	trans := &ioTransport{}
	trans.reader = reader
	trans.writer = writerCloser
	trans.closer = writerCloser
	return trans
}

func NewCloseableMaster(exec Executor, closer io.Closer) CloseableMaster {
	master := &closableMaster{}
	master.exec = exec
	master.closer = closer.Close
	return master
}

func NewTransportExecutor(proto Protocol, trans Transport, toms int) CloseableExecutor {
	exec := &transportExecutor{}
	exec.trans = trans
	exec.proto = proto
	exec.toms = toms
	return exec
}

func NewModelExecutor(model Model) Executor {
	exec := &modelExecutor{}
	exec.model = model
	return exec
}

func NewMapModel() *mapModel {
	m := &mapModel{}
	m.dis = make(map[string]bool)
	m.dos = make(map[string]bool)
	m.wis = make(map[string]uint16)
	m.wos = make(map[string]uint16)
	return m
}

const (
	ReadDos01  byte = 1
	ReadDis02  byte = 2
	ReadWos03  byte = 3
	ReadWis04  byte = 4
	WriteDo05  byte = 5
	WriteWo06  byte = 6
	WriteDos15 byte = 15
	WriteWos16 byte = 16
	MaxBools        = 255 * 8
	MaxWords        = 255 / 2
	TrueWord        = 0xFF00
	ReadToMs        = 100
)

type Command struct {
	Slave   byte
	Code    byte
	Address uint16
	Corv    uint16 //count or value
	Bools   []bool
	Words   []uint16
}

type Executor interface {
	Execute(c *Command) (*Command, error)
}

type Master interface {
	ReadDo(slave byte, address uint16) (bool, error)
	ReadDi(slave byte, address uint16) (bool, error)
	ReadWi(slave byte, address uint16) (uint16, error)
	ReadWo(slave byte, address uint16) (uint16, error)
	ReadDos(slave byte, address uint16, count uint16) ([]bool, error)
	ReadDis(slave byte, address uint16, count uint16) ([]bool, error)
	ReadWis(slave byte, address uint16, count uint16) ([]uint16, error)
	ReadWos(slave byte, address uint16, count uint16) ([]uint16, error)
	WriteDo(slave byte, address uint16, value bool) error
	WriteWo(slave byte, address uint16, value uint16) error
	WriteDos(slave byte, address uint16, values ...bool) error
	WriteWos(slave byte, address uint16, values ...uint16) error
}

type Model interface {
	ReadDis(slave byte, address uint16, count uint16) []bool
	ReadDos(slave byte, address uint16, count uint16) []bool
	ReadWis(slave byte, address uint16, count uint16) []uint16
	ReadWos(slave byte, address uint16, count uint16) []uint16
	WriteDis(slave byte, address uint16, values ...bool)
	WriteDos(slave byte, address uint16, values ...bool)
	WriteWis(slave byte, address uint16, values ...uint16)
	WriteWos(slave byte, address uint16, values ...uint16)
}

type Protocol interface {
	CheckWrapper(buf []byte, length uint16) error
	MakeBuffers(length uint16) ([]byte, []byte)
	WrapBuffer(buf []byte, length uint16)
	Scan(t Transport) (*Command, error)
	ExceptionLen() int
	Finally()
}

type Transport interface {
	io.Writer
	io.Closer
	//internally applied only after an error was reported
	//should only report io.EOF errors
	Discard() error
	SetError(eflag bool)
	//expected to return partial read on timeout to detect exception
	//expected to never return a negative counter
	//toms<0 returns only on full buffer or io.EOF
	//toms=0 returns data readily available
	//toms>0 reads as much as possible until expiration
	//must detect breaks at middle of packet
	TimedRead(buf []byte, toms int) (int, error)
}

type TimedReader interface {
	//expected to never return a negative counter
	//must timeout at ReadToMs constant value
	TimedRead(buf []byte) (int, error)
}

type CloseableMaster interface {
	io.Closer

	Master
}

type CloseableExecutor interface {
	io.Closer

	Executor
}
