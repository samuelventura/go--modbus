package spec

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"runtime/debug"
	"testing"
	"time"

	"github.com/samuelventura/go-modbus"
)

func ProtocolTest(t *testing.T, proto modbus.Protocol, setup func(proto modbus.Protocol) (model modbus.Model, master modbus.CloseableMaster, err error)) {
	log.Println("protocol", reflect.TypeOf(proto))
	model, master, err := setup(proto)
	fatalIfError(t, err)
	testModelMaster(t, model, master)
}

//MASTER////////////////////////////

func testModelMaster(t *testing.T, model modbus.Model, master modbus.CloseableMaster) {
	defer master.Close()
	var bools []bool
	var words []uint16
	var bool1 bool
	var word1 uint16
	var err error

	testWriteDos(t, model, master, 0, 0, randBools(modbus.MaxBools)...)
	testWriteWos(t, model, master, 0, 0, randWords(modbus.MaxWords)...)
	testReadDos(t, model, master, 0, 0, randBools(modbus.MaxBools)...)
	testReadWos(t, model, master, 0, 0, randWords(modbus.MaxWords)...)
	testReadDis(t, model, master, 0, 0, randBools(modbus.MaxBools)...)
	testReadWis(t, model, master, 0, 0, randWords(modbus.MaxWords)...)

	for k := 0; k < 10; k++ {
		testWriteDos(t, model, master, 0, 0, randBools(modbus.MaxBools-k)...)
		testWriteWos(t, model, master, 0, 0, randWords(modbus.MaxWords-k)...)
		testReadDos(t, model, master, 0, 0, randBools(modbus.MaxBools-k)...)
		testReadWos(t, model, master, 0, 0, randWords(modbus.MaxWords-k)...)
		testReadDis(t, model, master, 0, 0, randBools(modbus.MaxBools-k)...)
		testReadWis(t, model, master, 0, 0, randWords(modbus.MaxWords-k)...)
	}

	err = master.WriteDo(0xFF, 0xFFFF, false)
	if err.Error() != fmt.Sprintf("modbus exception %02x", ^modbus.WriteDo05) {
		t.Fatalf("exception expected: %s", err.Error())
	}
	if me, ok := err.(*modbus.ModbusException); !ok || me.Code != ^modbus.WriteDo05 {
		t.Fatalf("exception expected: %s", err.Error())
	}
	max := 0x10001
	start := time.Now().UnixNano()
	for k := 0; k < max; k++ {
		fatalIfError(t, master.WriteDo(0, 0, false))
	}
	end := time.Now().UnixNano()
	totals := float64(end-start) / 1000000000.0
	unitms := float64(end-start) / float64(max) / 1000000.0
	log.Printf("Timed %fs %fms %d\n", totals, unitms, max)
	for ss := 0; ss < 0x1FF; ss += 50 {
		for aa := 0; aa < 0x1FFFF; aa += 10000 {
			s := byte(ss)
			a := uint16(aa)

			fatalIfError(t, master.WriteDo(s, a, true))
			assertBoolsEqual(t, model.ReadDos(s, a, 1), []bool{true})
			bools, err = master.ReadDos(s, a, 1)
			assertBoolsEqualErr(t, err, bools, []bool{true})
			bool1, err = master.ReadDo(s, a)
			assertBoolEqualErr(t, err, bool1, true)
			fatalIfError(t, master.WriteDo(s, a, false))
			assertBoolsEqual(t, model.ReadDos(s, a, 1), []bool{false})
			bools, err = master.ReadDos(s, a, 1)
			assertBoolsEqualErr(t, err, bools, []bool{false})
			bool1, err = master.ReadDo(s, a)
			assertBoolEqualErr(t, err, bool1, false)

			fatalIfError(t, master.WriteWo(s, a, 0x37A5))
			assertWordsEqual(t, model.ReadWos(s, a, 1), []uint16{0x37A5})
			words, err = master.ReadWos(s, a, 1)
			assertWordsEqualErr(t, err, words, []uint16{0x37A5})
			word1, err = master.ReadWo(s, a)
			assertWordEqualErr(t, err, word1, 0x37A5)
			fatalIfError(t, master.WriteWo(s, a, 0xC8F0))
			assertWordsEqual(t, model.ReadWos(s, a, 1), []uint16{0xC8F0})
			words, err = master.ReadWos(s, a, 1)
			assertWordsEqualErr(t, err, words, []uint16{0xC8F0})
			word1, err = master.ReadWo(s, a)
			assertWordEqualErr(t, err, word1, 0xC8F0)

			a += 1
			fatalIfError(t, master.WriteDos(s, a, true, true))
			assertBoolsEqual(t, model.ReadDos(s, a, 2), []bool{true, true})
			fatalIfError(t, master.WriteDos(s, a, false, true))
			assertBoolsEqual(t, model.ReadDos(s, a, 2), []bool{false, true})
			fatalIfError(t, master.WriteDos(s, a, true, false))
			assertBoolsEqual(t, model.ReadDos(s, a, 2), []bool{true, false})
			fatalIfError(t, master.WriteDos(s, a, false, false))
			assertBoolsEqual(t, model.ReadDos(s, a, 2), []bool{false, false})

			fatalIfError(t, master.WriteWos(s, a, 0x37A5, 0xC8F0))
			assertWordsEqual(t, model.ReadWos(s, a, 2), []uint16{0x37A5, 0xC8F0})
			fatalIfError(t, master.WriteWos(s, a, 0xC80F, 0x37A5))
			assertWordsEqual(t, model.ReadWos(s, a, 2), []uint16{0xC80F, 0x37A5})

			a += 2
			model.WriteDis(s, a, true, true)
			bools, err = master.ReadDis(s, a, 2)
			assertBoolsEqualErr(t, err, bools, []bool{true, true})
			bool1, err = master.ReadDi(s, a)
			assertBoolEqualErr(t, err, bool1, true)
			bool1, err = master.ReadDi(s, a+1)
			assertBoolEqualErr(t, err, bool1, true)
			model.WriteDis(s, a, false, true)
			bools, err = master.ReadDis(s, a, 2)
			assertBoolsEqualErr(t, err, bools, []bool{false, true})
			bool1, err = master.ReadDi(s, a)
			assertBoolEqualErr(t, err, bool1, false)
			bool1, err = master.ReadDi(s, a+1)
			assertBoolEqualErr(t, err, bool1, true)
			model.WriteDis(s, a, true, false)
			bools, err = master.ReadDis(s, a, 2)
			assertBoolsEqualErr(t, err, bools, []bool{true, false})
			bool1, err = master.ReadDi(s, a)
			assertBoolEqualErr(t, err, bool1, true)
			bool1, err = master.ReadDi(s, a+1)
			assertBoolEqualErr(t, err, bool1, false)
			model.WriteDis(s, a, false, false)
			bools, err = master.ReadDis(s, a, 2)
			assertBoolsEqualErr(t, err, bools, []bool{false, false})
			bool1, err = master.ReadDi(s, a)
			assertBoolEqualErr(t, err, bool1, false)
			bool1, err = master.ReadDi(s, a+1)
			assertBoolEqualErr(t, err, bool1, false)

			a += 2
			model.WriteWis(s, a, 0x37A5, 0xC8F0)
			words, err = master.ReadWis(s, a, 2)
			assertWordsEqualErr(t, err, words, []uint16{0x37A5, 0xC8F0})
			word1, err = master.ReadWi(s, a)
			assertWordEqualErr(t, err, word1, 0x37A5)
			word1, err = master.ReadWi(s, a+1)
			assertWordEqualErr(t, err, word1, 0xC8F0)
			model.WriteWis(s, a, 0xC80F, 0x375A)
			words, err = master.ReadWis(s, a, 2)
			assertWordsEqualErr(t, err, words, []uint16{0xC80F, 0x375A})
			word1, err = master.ReadWi(s, a)
			assertWordEqualErr(t, err, word1, 0xC80F)
			word1, err = master.ReadWi(s, a+1)
			assertWordEqualErr(t, err, word1, 0x375A)

			a += 2
			testBools(t, model, master, s, a, true, false, true, false, true, false, true, false, true, true, true, true, false, false, false, false, true)
			testBools(t, model, master, s, a, true, true, true, true, false, false, false, false, true, true, false, true, false, true, false, true, false)
			testWords(t, model, master, s, a, 0x0102, 0x0304, 0x0506, 0x0708, 0x09A0, 0x5A73, 0x000, 0xFFFF, 0xDE45, 0x98FE, 0x00FF, 0xFF00, 0x000, 0xFFFF)

			bools = randBools(20)
			words = randWords(20)
			for j := 1; j <= 20; j++ {
				testBools(t, model, master, s, a, bools[:j]...)
				testWords(t, model, master, s, a, words[:j]...)
			}
		}
	}
}

func testWriteDos(t *testing.T, model modbus.Model, master modbus.Master, s byte, a uint16, values ...bool) {
	fatalIfError(t, master.WriteDos(s, a, values...))
	bools := model.ReadDos(s, a, uint16(len(values)))
	assertBoolsEqual(t, values, bools)
}

func testWriteWos(t *testing.T, model modbus.Model, master modbus.Master, s byte, a uint16, values ...uint16) {
	fatalIfError(t, master.WriteWos(s, a, values...))
	words := model.ReadWos(s, a, uint16(len(values)))
	assertWordsEqual(t, values, words)
}

func testReadDos(t *testing.T, model modbus.Model, master modbus.Master, s byte, a uint16, values ...bool) {
	model.WriteDos(s, a, values...)
	bools, err := master.ReadDos(s, a, uint16(len(values)))
	assertBoolsEqualErr(t, err, values, bools)
}

func testReadWos(t *testing.T, model modbus.Model, master modbus.Master, s byte, a uint16, values ...uint16) {
	model.WriteWos(s, a, values...)
	words, err := master.ReadWos(s, a, uint16(len(values)))
	assertWordsEqualErr(t, err, values, words)
}

func testReadDis(t *testing.T, model modbus.Model, master modbus.Master, s byte, a uint16, values ...bool) {
	model.WriteDis(s, a, values...)
	bools, err := master.ReadDis(s, a, uint16(len(values)))
	assertBoolsEqualErr(t, err, values, bools)
}

func testReadWis(t *testing.T, model modbus.Model, master modbus.Master, s byte, a uint16, values ...uint16) {
	model.WriteWis(s, a, values...)
	words, err := master.ReadWis(s, a, uint16(len(values)))
	assertWordsEqualErr(t, err, values, words)
}

func testBools(t *testing.T, model modbus.Model, master modbus.Master, s byte, a uint16, values ...bool) {
	var err error
	var bools []bool
	model.WriteDis(s, a, values...)
	bools, err = master.ReadDis(s, a, uint16(len(values)))
	assertBoolsEqualErr(t, err, values, bools)
	model.WriteDos(s, a, values...)
	bools, err = master.ReadDos(s, a, uint16(len(values)))
	assertBoolsEqualErr(t, err, values, bools)
	master.WriteDos(s, a, values...)
	assertBoolsEqual(t, values, model.ReadDos(s, a, uint16(len(values))))
}

func testWords(t *testing.T, model modbus.Model, master modbus.Master, s byte, a uint16, values ...uint16) {
	var err error
	var words []uint16
	model.WriteWis(s, a, values...)
	words, err = master.ReadWis(s, a, uint16(len(values)))
	assertWordsEqualErr(t, err, values, words)
	model.WriteWos(s, a, values...)
	words, err = master.ReadWos(s, a, uint16(len(values)))
	assertWordsEqualErr(t, err, values, words)
	master.WriteWos(s, a, values...)
	assertWordsEqual(t, values, model.ReadWos(s, a, uint16(len(values))))
}

//SLAVE////////////////////////////

type ExceptionExecutor struct {
	Exec modbus.Executor
}

func (e *ExceptionExecutor) Execute(ci *modbus.Command) (co *modbus.Command, err error) {
	if ci.Slave == 0xFF && ci.Address == 0xFFFF {
		err = formatErr("Exception")
		return
	}
	return e.Exec.Execute(ci)
}

func setupMasterSlave(proto modbus.Protocol) (model modbus.Model, master modbus.CloseableMaster, err error) {
	listen, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Panic(err)
	}
	port := listen.Addr().(*net.TCPAddr).Port
	model = modbus.NewMapModel()
	exec := modbus.NewModelExecutor(model)
	execw := &ExceptionExecutor{exec}
	go func() {
		defer listen.Close()
		input, err := listen.Accept()
		if err != nil {
			log.Println("accept failed", err)
			return
		}
		itrans := modbus.NewConnTransport(input)
		modbus.RunSlave(proto, itrans, execw)
	}()
	otrans, err := modbus.NewTcpTransport(fmt.Sprintf(":%d", port), 0)
	if err != nil {
		return
	}
	master = modbus.NewMaster(proto, otrans, 400)
	return
}

//ASSERT////////////////////////////

func assertBoolEqualErr(t *testing.T, err error, a, b bool) {
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("val mismatch %t %t", a, b)
	}
}

func assertBoolsEqualErr(t *testing.T, err error, a, b []bool) {
	if err != nil {
		t.Fatal(err)
	}
	assertBoolsEqual(t, a, b)
}

func assertBoolsEqual(t *testing.T, a, b []bool) {
	if len(a) != len(b) {
		t.Fatalf("len mismatch %d %d", len(a), len(b))
	}
	for i, v := range a {
		if v != b[i] {
			t.Fatalf("val mismatch at %d %t %t", i, v, b[i])
		}
	}
}

func assertWordEqualErr(t *testing.T, err error, a, b uint16) {
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("val mismatch %04x %04x", a, b)
	}
}

func assertWordsEqualErr(t *testing.T, err error, a, b []uint16) {
	if err != nil {
		t.Fatal(err)
	}
	assertWordsEqual(t, a, b)
}

func assertWordsEqual(t *testing.T, a, b []uint16) {
	if len(a) != len(b) {
		t.Fatalf("len mismatch %d %d", len(a), len(b))
	}
	for i, v := range a {
		if v != b[i] {
			t.Fatalf("val mismatch at %d %04x %04x", i, v, b[i])
		}
	}
}

//RAND/////////////////////////////////////

func randBools(count int) (bools []bool) {
	bools = make([]bool, count)
	for i := range bools {
		bools[i] = randBool()
	}
	return
}

func randWords(count int) (words []uint16) {
	words = make([]uint16, count)
	for i := range words {
		words[i] = randWord()
	}
	return
}

func randBool() bool {
	return time.Now().UnixNano()%2 == 1
}

func randWord() uint16 {
	return uint16(time.Now().UnixNano() % 65536)
}

//TOOLS/////////////////////////////////////

func logPanic() {
	if r := recover(); r != nil {
		log.Println(r, string(debug.Stack()))
	}
}

func fatalIfError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func formatErr(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s %s", msg, string(debug.Stack()))
}
