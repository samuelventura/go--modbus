package modbus

import "fmt"

func (m *MapModel) Execute(ci *Command) (co *Command, err error) {
	co = &Command{}
	co.Slave = ci.Slave
	co.Code = ci.Code
	co.Address = ci.Address
	co.Corv = ci.Corv
	switch ci.Code {
	case ReadDos01:
		co.Bools = m.ReadDos(ci.Slave, ci.Address, ci.Corv)
	case ReadDis02:
		co.Bools = m.ReadDis(ci.Slave, ci.Address, ci.Corv)
	case ReadWos03:
		co.Words = m.ReadWos(ci.Slave, ci.Address, ci.Corv)
	case ReadWis04:
		co.Words = m.ReadWis(ci.Slave, ci.Address, ci.Corv)
	case WriteDo05:
		m.WriteDos(ci.Slave, ci.Address, ci.Corv == TrueWord)
	case WriteWo06:
		m.WriteWos(ci.Slave, ci.Address, ci.Corv)
	case WriteDos15:
		m.WriteDos(ci.Slave, ci.Address, ci.Bools...)
	case WriteWos16:
		m.WriteWos(ci.Slave, ci.Address, ci.Words...)
	default:
		err = formatErr("unsupported code %d", ci.Code)
		return
	}
	return
}

func (m *MapModel) ReadDis(slave byte, address uint16, count uint16) []bool {
	a := int(address)
	values := make([]bool, count)
	for i := range values {
		values[i] = m.dis[m.Key(slave, a+i)]
	}
	return values
}

func (m *MapModel) ReadDos(slave byte, address uint16, count uint16) []bool {
	a := int(address)
	values := make([]bool, count)
	for i := range values {
		values[i] = m.dos[m.Key(slave, a+i)]
	}
	return values
}

func (m *MapModel) WriteDis(slave byte, address uint16, values ...bool) {
	a := int(address)
	for i := range values {
		m.dis[m.Key(slave, a+i)] = values[i]
	}
}

func (m *MapModel) WriteDos(slave byte, address uint16, values ...bool) {
	a := int(address)
	for i := range values {
		m.dos[m.Key(slave, a+i)] = values[i]
	}
}

func (m *MapModel) ReadWis(slave byte, address uint16, count uint16) []uint16 {
	a := int(address)
	values := make([]uint16, count)
	for i := range values {
		values[i] = m.wis[m.Key(slave, a+i)]
	}
	return values
}

func (m *MapModel) ReadWos(slave byte, address uint16, count uint16) []uint16 {
	a := int(address)
	values := make([]uint16, count)
	for i := range values {
		values[i] = m.wos[m.Key(slave, a+i)]
	}
	return values
}

func (m *MapModel) WriteWis(slave byte, address uint16, values ...uint16) {
	a := int(address)
	for i := range values {
		m.wis[m.Key(slave, a+i)] = values[i]
	}
}

func (m *MapModel) WriteWos(slave byte, address uint16, values ...uint16) {
	a := int(address)
	for i := range values {
		m.wos[m.Key(slave, a+i)] = values[i]
	}
}

func (m *MapModel) Key(slave byte, address int) string {
	return fmt.Sprintf("%d_%04x", slave, address)
}
