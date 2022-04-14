package serial

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adangel/nt5000-serial/emulator"
	"github.com/adangel/nt5000-serial/protocol"
	"go.bug.st/serial"
)

var port serial.Port = nil

func List() []string {
	ports, _ := serial.GetPortsList()
	return ports
}

func Connect(serialport string) {
	mode := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	var err error
	port, err = serial.Open(serialport, mode)
	if err != nil {
		log.Fatal(err)
	}
}

func Disconnect() {
	if port != nil {
		err := port.Close()
		if err != nil {
			log.Fatal(err)
		}
	}
	port = nil
}

func isConnected() {
	if port == nil {
		log.Fatal("Not connected")
	}
}

func Send(data []byte) {
	isConnected()

	n, err := port.Write(data)
	if err != nil {
		log.Fatal(err)
	}
	if n != len(data) {
		log.Fatalf("Couldn't send all bytes, only %v of %v bytes sent\n", n, len(data))
	}

	log.Printf("Sent %v bytes: %x\n", n, data)
}

func Receive() ([]byte, error) {
	isConnected()

	port.SetReadTimeout(time.Millisecond * 250)

	result := make([]byte, 0, 26)

	for {
		readbuff := make([]byte, 13)
		n, err := port.Read(readbuff)
		if err != nil {
			log.Fatal(err)
			break
		}
		if n == 0 {
			log.Printf("Timeout after %v bytes\n", len(result))
			break
		}
		result = append(result, readbuff[:n]...)
		log.Printf("Received %v bytes (0x%x)\n", n, readbuff[:n])
	}

	var err error = nil
	if len(result) == 0 {
		err = fmt.Errorf("Didn't receive any data\n")
	} else {
		log.Printf("Received %v bytes: 0x%x\n", len(result), result)
	}
	return result, err
}

// see https://golangcode.com/handle-ctrl-c-exit-in-terminal/
func SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Printf("Ctlr+C pressed, exiting...")
		Disconnect()
		os.Exit(0)
	}()
}

func GetDataPoint(emulate bool) protocol.DataPoint {
	var err error
	buff := make([]byte, 13)

	if emulate {
		buff = protocol.ConvertToByte(emulator.ProduceDataPoint())
	} else {
		Send([]byte("\x00\x01\x02\x01\x04"))
		buff, err = Receive()
		if err != nil {
			log.Print(err)
		}
	}

	err = protocol.VerifyChecksum(buff)
	if err != nil {
		log.Print(err)
	}

	d, err := protocol.Convert(buff)
	if err != nil {
		log.Printf("Invalid data received: %v\n", err)
	}
	return d
}

func ReadSerialNumber(emulate bool) string {
	var err error
	buff := make([]byte, 13)

	if emulate {
		buff = []byte("1533A5012345\x71")
	} else {
		Send([]byte("\x00\x01\x08\x01\x0A"))
		buff, err = Receive()
		if err != nil {
			log.Print(err)
		}
	}

	err = protocol.VerifyChecksum(buff)
	if err != nil {
		log.Print(err)
	}

	var serialnumber string = ""
	for i := 0; i < 12; i++ {
		if buff[i] != 0x0d {
			serialnumber += string(buff[i])
		}
	}
	return serialnumber
}

func ReadProtocolAndFirmware(emulate bool) (string, string) {
	var err error
	buff := make([]byte, 13)

	if emulate {
		buff = []byte("111-23\x00\x00\x00\x00\x00\x00\x25")
	} else {
		Send([]byte("\x00\x01\x09\x01\x0B"))
		buff, err = Receive()
		if err != nil {
			log.Print(err)
		}
	}

	err = protocol.VerifyChecksum(buff)
	if err != nil {
		log.Print(err)
	}

	var protocol string = string(buff[0:2])
	var firmware string = ""

	for i := 2; i < 11; i++ {
		if buff[i] != 0x0d {
			firmware += string(buff[i])
		}
	}

	return protocol, firmware
}

func ReadErrors(emulate bool) []protocol.Error {
	var result []protocol.Error = make([]protocol.Error, 0, 10)

	for i := 1; i <= 5; i++ {
		result = append(result, readSingleError(emulate, uint8(i))...)
	}
	return result
}

func readSingleError(emulate bool, slot uint8) []protocol.Error {
	var err error
	buff := make([]byte, 13)
	if emulate {
		for i := 0; i < 12; i++ {
			buff[i] = 0x0d
		}
		if slot == 1 {
			now := time.Now().Local()
			buff[0] = byte(now.Month())
			buff[1] = byte(now.Day())
			buff[2] = 20   // hour
			buff[3] = 3    // minute
			buff[4] = 0x11 // error code
			buff[5] = byte(now.Year() - 2000)
		}
		protocol.CalculateChecksum(buff)
	} else {
		req := []byte("\x00\x01\x01")
		req = append(req, byte(slot), 0x00)
		protocol.CalculateChecksum(req)
		Send(req)
		buff, err = Receive()
		if err != nil {
			log.Print(err)
		}
	}

	err = protocol.VerifyChecksum(buff)
	if err != nil {
		log.Print(err)
	}

	var result []protocol.Error
	if buff[0] != 0x0d {
		result = append(result, protocol.Error{
			Date: time.Date(int(buff[5])+2000, time.Month(buff[0]), int(buff[1]), int(buff[2]), int(buff[3]), 0, 0, time.Local),
			Code: buff[4],
		})
	}
	if buff[6] != 0x0d {
		result = append(result, protocol.Error{
			Date: time.Date(int(buff[11])+2000, time.Month(buff[6]), int(buff[7]), int(buff[8]), int(buff[9]), 0, 0, time.Local),
			Code: buff[10],
		})
	}
	return result
}
