package modbus

import (
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

func TestModbus(t *testing.T) {
	defer logPanic()
	log.SetFlags(log.Lmicroseconds)
	//traceEnabled = true
	//EnableTrace(true)
	testProtocol(t, NewNopProtocol())
	testProtocol(t, NewRtuProtocol())
	testProtocol(t, NewTcpProtocol())
}

func testProtocol(t *testing.T, proto Protocol) {
	log.Println("protocol", reflect.TypeOf(proto))
	model, master, err := setupMasterSlave(proto)
	ifErrFatal(t, err)
	testModelMaster(t, model, master)
}

//MASTER////////////////////////////

func testModelMaster(t *testing.T, model Model, master CloseableMaster) {
	defer master.Close()
	var bools []bool
	var words []uint16
	var bool1 bool
	var word1 uint16
	var err error

	testWriteDos(t, model, master, 0, 0, randBools(MaxBools)...)
	testWriteWos(t, model, master, 0, 0, randWords(MaxWords)...)
	testReadDos(t, model, master, 0, 0, randBools(MaxBools)...)
	testReadWos(t, model, master, 0, 0, randWords(MaxWords)...)
	testReadDis(t, model, master, 0, 0, randBools(MaxBools)...)
	testReadWis(t, model, master, 0, 0, randWords(MaxWords)...)

	for k := 0; k < 10; k++ {
		testWriteDos(t, model, master, 0, 0, randBools(MaxBools-k)...)
		testWriteWos(t, model, master, 0, 0, randWords(MaxWords-k)...)
		testReadDos(t, model, master, 0, 0, randBools(MaxBools-k)...)
		testReadWos(t, model, master, 0, 0, randWords(MaxWords-k)...)
		testReadDis(t, model, master, 0, 0, randBools(MaxBools-k)...)
		testReadWis(t, model, master, 0, 0, randWords(MaxWords-k)...)
	}

	err = master.WriteDo(0xFF, 0xFFFF, false)
	if !strings.HasPrefix(err.Error(), fmt.Sprintf("modbusException %02x", ^WriteDo05)) {
		t.Fatalf("exception expected: %s", err.Error())
	}
	max := 0x10001
	start := time.Now().UnixNano()
	for k := 0; k < max; k++ {
		ifErrFatal(t, master.WriteDo(0, 0, false))
	}
	end := time.Now().UnixNano()
	totals := float64(end-start) / 1000000000.0
	unitms := float64(end-start) / float64(max) / 1000000.0
	log.Printf("Timed %fs %fms %d\n", totals, unitms, max)
	for ss := 0; ss < 0x1FF; ss += 50 {
		for aa := 0; aa < 0x1FFFF; aa += 10000 {
			s := byte(ss)
			a := uint16(aa)

			ifErrFatal(t, master.WriteDo(s, a, true))
			assertBoolsEqual(t, model.ReadDos(s, a, 1), []bool{true})
			bools, err = master.ReadDos(s, a, 1)
			assertBoolsEqualErr(t, err, bools, []bool{true})
			bool1, err = master.ReadDo(s, a)
			assertBoolEqualErr(t, err, bool1, true)
			ifErrFatal(t, master.WriteDo(s, a, false))
			assertBoolsEqual(t, model.ReadDos(s, a, 1), []bool{false})
			bools, err = master.ReadDos(s, a, 1)
			assertBoolsEqualErr(t, err, bools, []bool{false})
			bool1, err = master.ReadDo(s, a)
			assertBoolEqualErr(t, err, bool1, false)

			ifErrFatal(t, master.WriteWo(s, a, 0x37A5))
			assertWordsEqual(t, model.ReadWos(s, a, 1), []uint16{0x37A5})
			words, err = master.ReadWos(s, a, 1)
			assertWordsEqualErr(t, err, words, []uint16{0x37A5})
			word1, err = master.ReadWo(s, a)
			assertWordEqualErr(t, err, word1, 0x37A5)
			ifErrFatal(t, master.WriteWo(s, a, 0xC8F0))
			assertWordsEqual(t, model.ReadWos(s, a, 1), []uint16{0xC8F0})
			words, err = master.ReadWos(s, a, 1)
			assertWordsEqualErr(t, err, words, []uint16{0xC8F0})
			word1, err = master.ReadWo(s, a)
			assertWordEqualErr(t, err, word1, 0xC8F0)

			a += 1
			ifErrFatal(t, master.WriteDos(s, a, true, true))
			assertBoolsEqual(t, model.ReadDos(s, a, 2), []bool{true, true})
			ifErrFatal(t, master.WriteDos(s, a, false, true))
			assertBoolsEqual(t, model.ReadDos(s, a, 2), []bool{false, true})
			ifErrFatal(t, master.WriteDos(s, a, true, false))
			assertBoolsEqual(t, model.ReadDos(s, a, 2), []bool{true, false})
			ifErrFatal(t, master.WriteDos(s, a, false, false))
			assertBoolsEqual(t, model.ReadDos(s, a, 2), []bool{false, false})

			ifErrFatal(t, master.WriteWos(s, a, 0x37A5, 0xC8F0))
			assertWordsEqual(t, model.ReadWos(s, a, 2), []uint16{0x37A5, 0xC8F0})
			ifErrFatal(t, master.WriteWos(s, a, 0xC80F, 0x37A5))
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

func testWriteDos(t *testing.T, model Model, master Master, s byte, a uint16, values ...bool) {
	ifErrFatal(t, master.WriteDos(s, a, values...))
	bools := model.ReadDos(s, a, uint16(len(values)))
	assertBoolsEqual(t, values, bools)
}

func testWriteWos(t *testing.T, model Model, master Master, s byte, a uint16, values ...uint16) {
	ifErrFatal(t, master.WriteWos(s, a, values...))
	words := model.ReadWos(s, a, uint16(len(values)))
	assertWordsEqual(t, values, words)
}

func testReadDos(t *testing.T, model Model, master Master, s byte, a uint16, values ...bool) {
	model.WriteDos(s, a, values...)
	bools, err := master.ReadDos(s, a, uint16(len(values)))
	assertBoolsEqualErr(t, err, values, bools)
}

func testReadWos(t *testing.T, model Model, master Master, s byte, a uint16, values ...uint16) {
	model.WriteWos(s, a, values...)
	words, err := master.ReadWos(s, a, uint16(len(values)))
	assertWordsEqualErr(t, err, values, words)
}

func testReadDis(t *testing.T, model Model, master Master, s byte, a uint16, values ...bool) {
	model.WriteDis(s, a, values...)
	bools, err := master.ReadDis(s, a, uint16(len(values)))
	assertBoolsEqualErr(t, err, values, bools)
}

func testReadWis(t *testing.T, model Model, master Master, s byte, a uint16, values ...uint16) {
	model.WriteWis(s, a, values...)
	words, err := master.ReadWis(s, a, uint16(len(values)))
	assertWordsEqualErr(t, err, values, words)
}

func testBools(t *testing.T, model Model, master Master, s byte, a uint16, values ...bool) {
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

func testWords(t *testing.T, model Model, master Master, s byte, a uint16, values ...uint16) {
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

func runSlave(proto Protocol, trans Transport, exec Executor) {
	for {
		err := oneSlave(proto, trans, exec)
		if err != nil {
			trace("oneSlave.error", err)
		}
		if err == io.EOF {
			return
		}
	}
}

func oneSlave(proto Protocol, trans Transport, exec Executor) (err error) {
	//report error to transport
	//to discard on next interaction
	defer func() {
		if err != nil {
			trans.DiscardOn()
		}
	}()
	trans.Discard()
	for {
		ci, err := proto.Scan(trans)
		if err != nil {
			return err
		}
		if ci.Slave == 0xFF && ci.Address == 0xFFFF {
			fbuf, buf := proto.MakeBuffers(3)
			buf[0] = ci.Slave
			buf[1] = ci.Code | 0x80
			buf[2] = ^ci.Code
			proto.WrapBuffer(fbuf, 3)
			trans.Write(fbuf)
			continue
		}
		_, buf, err := applyToExecutor(ci, proto, exec)
		if err != nil {
			return err
		}
		c, err := trans.Write(buf)
		if err != nil {
			return err
		}
		if c != len(buf) {
			return formatErr("partial write %d of %d", c, len(buf))
		}
	}
}

func setupMasterSlave(proto Protocol) (model *mapModel, master CloseableMaster, err error) {
	listen, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Panic(err)
	}
	port := listen.Addr().(*net.TCPAddr).Port
	model = NewMapModel()
	exec := NewModelExecutor(model)
	go func() {
		defer listen.Close()
		input, err := listen.Accept()
		if err != nil {
			trace("accept failed", err)
			return
		}
		itrans := NewConnTransport(input)
		runSlave(proto, itrans, exec)
	}()
	otrans, err := NewTcpTransport(fmt.Sprintf(":%d", port), 0)
	if err != nil {
		return
	}
	master = NewMaster(proto, otrans, 400)
	return
}

func applyToExecutor(ci *Command, p Protocol, e Executor) (co *Command, fbuf []byte, err error) {
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

func ifErrFatal(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
