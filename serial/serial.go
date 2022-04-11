package serial

import (
	"fmt"
	"log"
	"time"

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
