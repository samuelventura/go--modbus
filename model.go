package modbus

import "fmt"

type mapModel struct {
	dis map[string]bool
	dos map[string]bool
	wis map[string]uint16
	wos map[string]uint16
}

func (m *mapModel) ReadDis(slave byte, address uint16, count uint16) []bool {
	a := int(address)
	values := make([]bool, count)
	for i := range values {
		values[i] = m.dis[m.Key(slave, a+i)]
	}
	return values
}

func (m *mapModel) ReadDos(slave byte, address uint16, count uint16) []bool {
	a := int(address)
	values := make([]bool, count)
	for i := range values {
		values[i] = m.dos[m.Key(slave, a+i)]
	}
	return values
}

func (m *mapModel) WriteDis(slave byte, address uint16, values ...bool) {
	a := int(address)
	for i := range values {
		m.dis[m.Key(slave, a+i)] = values[i]
	}
}

func (m *mapModel) WriteDos(slave byte, address uint16, values ...bool) {
	a := int(address)
	for i := range values {
		m.dos[m.Key(slave, a+i)] = values[i]
	}
}

func (m *mapModel) ReadWis(slave byte, address uint16, count uint16) []uint16 {
	a := int(address)
	values := make([]uint16, count)
	for i := range values {
		values[i] = m.wis[m.Key(slave, a+i)]
	}
	return values
}

func (m *mapModel) ReadWos(slave byte, address uint16, count uint16) []uint16 {
	a := int(address)
	values := make([]uint16, count)
	for i := range values {
		values[i] = m.wos[m.Key(slave, a+i)]
	}
	return values
}

func (m *mapModel) WriteWis(slave byte, address uint16, values ...uint16) {
	a := int(address)
	for i := range values {
		m.wis[m.Key(slave, a+i)] = values[i]
	}
}

func (m *mapModel) WriteWos(slave byte, address uint16, values ...uint16) {
	a := int(address)
	for i := range values {
		m.wos[m.Key(slave, a+i)] = values[i]
	}
}

func (m *mapModel) Key(slave byte, address int) string {
	return fmt.Sprintf("%d_%04x", slave, address)
}
