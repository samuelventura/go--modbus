package modbus

type transportExecutor struct {
	proto Protocol
	trans Transport
	close func() error
	toms  int
}

func (e *transportExecutor) Execute(ci *Command) (co *Command, err error) {
	co, err = ApplyToTransport(ci, e.proto, e.trans, e.toms)
	return
}

func (e *transportExecutor) Close() error {
	if e.close != nil {
		return e.close()
	}
	return nil
}
