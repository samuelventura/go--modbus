package main

import (
	"log"
	"time"

	"github.com/samuelventura/go-modbus"
)

//(cd sample; go run .)
func main() {
	trans, err := modbus.NewTcpTransport("10.77.0.10:502", 4000)
	if err != nil {
		log.Fatal(err)
	}
	defer trans.Close()
	modbus.EnableTrace(true)
	master := modbus.NewTcpMaster(trans, 400)
	for {
		err = master.WriteDo(1, 4, true)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(100 * time.Millisecond)
		err = master.WriteDo(1, 4, false)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
