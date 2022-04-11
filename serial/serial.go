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
	"github.com/spf13/cobra"
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

func ListSerialPorts(cmd *cobra.Command, args []string) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}
	fmt.Printf("Found %v ports:\n", len(ports))
	for _, port := range ports {
		fmt.Printf("- %v\n", port.Name)
		if port.IsUSB {
			fmt.Printf("   USB ID     %s:%s\n", port.VID, port.PID)
			fmt.Printf("   USB serial %s\n", port.SerialNumber)
			if len(port.Product) != 0 {
				fmt.Printf("   Product    %s\n", port.Product)
			}
		}
	}
}

var port serial.Port = nil

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

	log.Printf("Sent %v bytes\n", n)
}

func Receive(data []byte) error {
	isConnected()

	port.SetReadTimeout(time.Second * 1)

	index := 0

	for {
		readbuff := make([]byte, 13)
		n, err := port.Read(readbuff)
		if err != nil {
			log.Fatal(err)
			break
		}
		if n == 0 {
			break
		}
		log.Printf("Received %v bytes: %v\n", n, readbuff[:n])
		for i := 0; i < n; i++ {
			data[index] = readbuff[i]
			index++
		}
		if index == len(data) {
			log.Printf("Received complete response: %v\n", data)
			break
		}
	}

	var err error = nil
	if index != len(data) {
		err = fmt.Errorf("Didn't receive all data, only %v of %v bytes received\n", index, len(data))
	}
	return err
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
	buff := make([]byte, 13)

	if emulate {
		buff = protocol.ConvertToByte(emulator.ProduceDataPoint())
	} else {
		Send([]byte("\x00\x01\x02\x01\x04"))
		Receive(buff)
	}

	err := protocol.VerifyChecksum(buff)
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
	buff := make([]byte, 13)

	if emulate {
		buff = []byte("1533A5012345\x71")
	} else {
		Send([]byte("\x00\x01\x08\x01\x0A"))
		Receive(buff)
	}

	err := protocol.VerifyChecksum(buff)
	if err != nil {
		log.Print(err)
	}

	return string(buff[:12])
}

func ReadProtocol(emulate bool) string {
	buff := make([]byte, 13)

	if emulate {
		buff = []byte("111-23\x00\x00\x00\x00\x00\x00\x25")
	} else {
		Send([]byte("\x00\x01\x09\x01\x0B"))
		Receive(buff)
	}

	err := protocol.VerifyChecksum(buff)
	if err != nil {
		log.Print(err)
	}

	return string(buff[:6])
}
