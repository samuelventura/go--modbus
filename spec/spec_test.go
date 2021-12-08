package spec

import (
	"log"
	"testing"

	"github.com/samuelventura/go-modbus"
)

func TestModbus(t *testing.T) {
	defer logPanic()
	log.SetFlags(log.Lmicroseconds)
	ProtocolTest(t, modbus.NewNopProtocol(), setupMasterSlave)
	ProtocolTest(t, modbus.NewRtuProtocol(), setupMasterSlave)
	ProtocolTest(t, modbus.NewTcpProtocol(), setupMasterSlave)
}
