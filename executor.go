package modbus

import "io"

// Implementes: Executor
// Aplies commands to a model
type modelExecutor struct {
	model Model
}

func (e *modelExecutor) Execute(ci *Command) (co *Command, err error) {
	co = &Command{}
	co.Slave = ci.Slave
	co.Code = ci.Code
	co.Address = ci.Address
	co.Corv = ci.Corv
	switch ci.Code {
	case ReadDos01:
		co.Bools = e.model.ReadDos(ci.Slave, ci.Address, ci.Corv)
	case ReadDis02:
		co.Bools = e.model.ReadDis(ci.Slave, ci.Address, ci.Corv)
	case ReadWos03:
		co.Words = e.model.ReadWos(ci.Slave, ci.Address, ci.Corv)
	case ReadWis04:
		co.Words = e.model.ReadWis(ci.Slave, ci.Address, ci.Corv)
	case WriteDo05:
		e.model.WriteDos(ci.Slave, ci.Address, ci.Corv == TrueWord)
	case WriteWo06:
		e.model.WriteWos(ci.Slave, ci.Address, ci.Corv)
	case WriteDos15:
		e.model.WriteDos(ci.Slave, ci.Address, ci.Bools...)
	case WriteWos16:
		e.model.WriteWos(ci.Slave, ci.Address, ci.Words...)
	default:
		err = formatErr("unsupported code %d", ci.Code)
		return
	}
	return
}

// Implementes: Executor, ClosableExecutor
// Aplies commands to a transport
type transportExecutor struct {
	io.Closer
	proto Protocol
	trans Transport
	toms  int
}

func (e *transportExecutor) Close() error {
	return e.trans.Close()
}

func (e *transportExecutor) Execute(ci *Command) (co *Command, err error) {
	trace("t>", ci)
	err = ci.CheckValid()
	if err != nil {
		return
	}
	defer e.proto.Finally()
	reqlen := ci.RequestLength()
	freq, req := e.proto.MakeBuffers(reqlen)
	ci.EncodeRequest(req)
	e.proto.WrapBuffer(freq, reqlen)
	trace("t>", freq)
	//report error to transport
	//to discard on next interaction
	defer func() {
		if err != nil {
			e.trans.DiscardOn()
		}
	}()
	err = e.trans.DiscardIf()
	if err != nil {
		return
	}
	_write, err := e.trans.Write(freq)
	if err != nil {
		return
	}
	write := len(freq)
	if _write != write {
		err = formatErr("write mismatch got %d expected %d", _write, write)
		return
	}
	reslen := ci.ResponseLength()
	fres, res := e.proto.MakeBuffers(reslen)
	_read, err := e.trans.TimedRead(fres, e.toms)
	trace("t<", fres[:_read])
	if _read == e.proto.ExceptionLen() { //6+3
		err = e.proto.CheckWrapper(fres, 3)
		if err != nil {
			return
		}
		err = ci.CheckException(res)
		if err != nil {
			return
		}
		err = formatErr("modbusException %02x", res[2])
		return
	}
	if err != nil {
		return
	}
	read := len(fres)
	if _read != read {
		err = formatErr("read mismatch got %d expected %d", _read, read)
		return
	}
	err = e.proto.CheckWrapper(fres, reslen)
	if err != nil {
		return
	}
	err = ci.CheckResponse(res)
	if err != nil {
		return
	}
	co = &Command{}
	//not enough info in response packet to parse reads
	co.DecodeResponse(res, ci.Corv)
	trace("t<", co)
	return
}
