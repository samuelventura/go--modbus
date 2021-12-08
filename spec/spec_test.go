package spec

import (
	"log"
	"testing"

	"github.com/samuelventura/go-modbus"
)

func TestModbus(t *testing.T) {
	defer logPanic()
	log.SetFlags(log.Lmicroseconds)
	setupMasterSlave(t, modbus.NewNopProtocol(), ProtocolTest)
	setupMasterSlave(t, modbus.NewRtuProtocol(), ProtocolTest)
	setupMasterSlave(t, modbus.NewTcpProtocol(), ProtocolTest)
}
