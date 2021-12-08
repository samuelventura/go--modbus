package modbus

func ApplyToExecutor(ci *Command, p Protocol, e Executor) (co *Command, fbuf []byte, err error) {
	Trace("e>", ci)
	err = ci.CheckValid()
	if err != nil {
		return
	}
	co, err = e.Execute(ci)
	if err != nil {
		return
	}
	Trace("e<", co)
	reslen := ci.ResponseLength()
	fbuf, buf := p.MakeBuffers(reslen)
	co.EncodeResponse(buf)
	p.WrapBuffer(fbuf, reslen)
	return
}

func RunSlave(proto Protocol, trans Transport, exec Executor) (err error) {
	for {
		err = RunOneSlave(proto, trans, exec)
		if err != nil {
			return
		}
	}
}

func RunOneSlave(proto Protocol, trans Transport, exec Executor) (err error) {
	defer func() {
		if err != nil {
			trans.DiscardOn()
		}
	}()
	trans.DiscardIf()
	ci, err := proto.Scan(trans)
	if err != nil {
		return
	}
	_, rbuf, err := ApplyToExecutor(ci, proto, exec)
	if err != nil {
		fbuf, buf := proto.MakeBuffers(3)
		buf[0] = ci.Slave
		buf[1] = ci.Code | 0x80
		buf[2] = ^ci.Code
		proto.WrapBuffer(fbuf, 3)
		rbuf = fbuf
	}
	c, err := trans.Write(rbuf)
	if err != nil {
		return
	}
	if c != len(rbuf) {
		err = formatErr("partial write %d of %d", c, len(rbuf))
		return
	}
	return
}
