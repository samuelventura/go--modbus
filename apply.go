package modbus

func ApplyToTransport(ci *Command, p Protocol, t Transport, toms int) (co *Command, err error) {
	trace("t>", ci)
	err = ci.CheckValid()
	if err != nil {
		return
	}
	defer p.Finally()
	reqlen := ci.RequestLength()
	freq, req := p.MakeBuffers(reqlen)
	ci.EncodeRequest(req)
	p.WrapBuffer(freq, reqlen)
	trace("t>", freq)
	//report error to transport
	//to discard on next interaction
	defer func() {
		t.SetError(err != nil)
	}()
	err = t.Discard(100)
	if err != nil {
		return
	}
	_write, err := t.Write(freq)
	if err != nil {
		return
	}
	write := len(freq)
	if _write != write {
		err = formatErr("Write mismatch got %d expected %d", _write, write)
		return
	}
	reslen := ci.ResponseLength()
	fres, res := p.MakeBuffers(reslen)
	_read, err := t.TimedRead(fres, toms, 100)
	if _read == p.ExceptionLen() { //6+3
		err = p.CheckWrapper(fres, 3)
		if err != nil {
			return
		}
		err = ci.CheckException(res)
		if err != nil {
			return
		}
		err = formatErr("ModbusException %02x", res[2])
		return
	}
	if err != nil {
		return
	}
	trace("t<", fres[:_read])
	read := len(fres)
	if _read != read {
		err = formatErr("Read mismatch got %d expected %d", _read, read)
		return
	}
	err = p.CheckWrapper(fres, reslen)
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

func ApplyToExecutor(ci *Command, p Protocol, e Executor) (co *Command, fbuf []byte, err error) {
	trace("e>", ci)
	err = ci.CheckValid()
	if err != nil {
		return
	}
	co, err = e.Execute(ci)
	if err != nil {
		return
	}
	trace("e<", co)
	reslen := ci.ResponseLength()
	fbuf, buf := p.MakeBuffers(reslen)
	co.EncodeResponse(buf)
	p.WrapBuffer(fbuf, reslen)
	return
}
