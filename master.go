package modbus

type closableMaster struct {
	exec   Executor
	closer func() error
}

func (m *closableMaster) Close() error {
	if m.closer != nil {
		return m.closer()
	}
	return nil
}

func (m *closableMaster) Execute(c *Command) (*Command, error) {
	return m.exec.Execute(c)
}

func (m *closableMaster) ReadDo(slave byte, address uint16) (res bool, err error) {
	ci := &Command{Slave: slave, Code: ReadDos01, Address: address, Corv: 1}
	co, err := m.Execute(ci)
	if err == nil {
		res = co.Bools[0]
	}
	return
}

func (m *closableMaster) ReadDi(slave byte, address uint16) (res bool, err error) {
	ci := &Command{Slave: slave, Code: ReadDis02, Address: address, Corv: 1}
	co, err := m.Execute(ci)
	if err == nil {
		res = co.Bools[0]
	}
	return
}

func (m *closableMaster) ReadWi(slave byte, address uint16) (res uint16, err error) {
	ci := &Command{Slave: slave, Code: ReadWis04, Address: address, Corv: 1}
	co, err := m.Execute(ci)
	if err == nil {
		res = co.Words[0]
	}
	return
}

func (m *closableMaster) ReadWo(slave byte, address uint16) (res uint16, err error) {
	ci := &Command{Slave: slave, Code: ReadWos03, Address: address, Corv: 1}
	co, err := m.Execute(ci)
	if err == nil {
		res = co.Words[0]
	}
	return
}

func (m *closableMaster) ReadDos(slave byte, address uint16, count uint16) (res []bool, err error) {
	ci := &Command{Slave: slave, Code: ReadDos01, Address: address, Corv: count}
	co, err := m.Execute(ci)
	if err == nil {
		res = co.Bools
	}
	return
}

func (m *closableMaster) ReadDis(slave byte, address uint16, count uint16) (res []bool, err error) {
	ci := &Command{Slave: slave, Code: ReadDis02, Address: address, Corv: count}
	co, err := m.Execute(ci)
	if err == nil {
		res = co.Bools
	}
	return
}

func (m *closableMaster) ReadWis(slave byte, address uint16, count uint16) (res []uint16, err error) {
	ci := &Command{Slave: slave, Code: ReadWis04, Address: address, Corv: count}
	co, err := m.Execute(ci)
	if err == nil {
		res = co.Words
	}
	return
}

func (m *closableMaster) ReadWos(slave byte, address uint16, count uint16) (res []uint16, err error) {
	ci := &Command{Slave: slave, Code: ReadWos03, Address: address, Corv: count}
	co, err := m.Execute(ci)
	if err == nil {
		res = co.Words
	}
	return
}

func (m *closableMaster) WriteDo(slave byte, address uint16, value bool) (err error) {
	ci := &Command{Slave: slave, Code: WriteDo05, Address: address}
	if value {
		ci.Corv = TrueWord
	}
	_, err = m.Execute(ci)
	return
}

func (m *closableMaster) WriteWo(slave byte, address uint16, value uint16) (err error) {
	ci := &Command{Slave: slave, Code: WriteWo06, Address: address, Corv: value}
	_, err = m.Execute(ci)
	return
}

func (m *closableMaster) WriteDos(slave byte, address uint16, values ...bool) (err error) {
	ci := &Command{Slave: slave, Code: WriteDos15, Address: address, Corv: uint16(len(values)), Bools: values}
	_, err = m.Execute(ci)
	return
}

func (m *closableMaster) WriteWos(slave byte, address uint16, values ...uint16) (err error) {
	ci := &Command{Slave: slave, Code: WriteWos16, Address: address, Corv: uint16(len(values)), Words: values}
	_, err = m.Execute(ci)
	return
}
