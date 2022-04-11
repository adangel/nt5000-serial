package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atomicgo/cursor"
	"github.com/spf13/cobra"

	"github.com/adangel/nt5000-serial/protocol"
	serial "github.com/adangel/nt5000-serial/serial"
)

var port string
var serialport string
var emulate bool

var rootCmd = &cobra.Command{
	Use:     "nt5000-serial",
	Short:   "communicate with sunways nt5000 converter via rs232",
	Version: "1.0.0",
}

var cmdWeb = &cobra.Command{
	Use:   "web",
	Short: "start web server",
	Run:   startWebServer,
}

var cmdSerial = &cobra.Command{
	Use:   "serial",
	Short: "list serial ports",
	Run:   serial.ListSerialPorts,
}

var settime bool
var cmdDatetime = &cobra.Command{
	Use:   "datetime",
	Short: "Get or set the current time",
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Using serial port %s", serialport)

		if settime {
			now := time.Now().Local()
			log.Printf("Setting date to %s\n", now.Format(time.ANSIC))

			serial.Connect(serialport)

			buff := make([]byte, 5)
			buff[0] = 0
			buff[1] = 1

			// year
			buff[2] = 0x50
			buff[3] = byte(now.Year() - 2000)
			protocol.CalculateChecksum(buff)
			serial.Send(buff)

			// month
			buff[2] = 0x51
			buff[3] = byte(now.Month())
			protocol.CalculateChecksum(buff)
			serial.Send(buff)

			// day
			buff[2] = 0x52
			buff[3] = byte(now.Day())
			protocol.CalculateChecksum(buff)
			serial.Send(buff)

			// hour
			buff[2] = 0x53
			buff[3] = byte(now.Hour())
			protocol.CalculateChecksum(buff)
			serial.Send(buff)

			// minute
			buff[2] = 0x54
			buff[3] = byte(now.Minute())
			protocol.CalculateChecksum(buff)
			serial.Send(buff)

			serial.Disconnect()
		} else {
			buff := make([]byte, 13)

			log.Println("Reading current date...")
			if emulate {
				now := time.Now().Local()
				buff[0] = byte(now.Year() - 2000)
				buff[1] = byte(now.Month())
				buff[2] = byte(now.Day())
				buff[3] = byte(now.Hour())
				buff[4] = byte(now.Minute())
				protocol.CalculateChecksum(buff)
			} else {
				serial.Connect(serialport)
				serial.Send([]byte("\x00\x01\x06\x01\x08"))
				serial.Receive(buff)
				serial.Disconnect()
			}

			err := protocol.VerifyChecksum(buff)
			if err != nil {
				log.Print(err)
			}

			t := time.Date(int(buff[0])+2000, time.Month(buff[1]), int(buff[2]), int(buff[3]), int(buff[4]), 0, 0, time.Local)
			fmt.Printf("Current time: %s\n", t.Local().Format(time.ANSIC))
		}
	},
}

var pollInterval uint8
var cmdDisplay = &cobra.Command{
	Use:   "display",
	Short: "Display current reading on the command line",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Displaying...")

		if pollInterval < 1 || pollInterval > 100 {
			fmt.Printf("Invalid poll interval %v specified, using default\n", pollInterval)
			pollInterval = 5
		}

		setupCloseHandler()

		area := cursor.NewArea()
		area.Clear()

		if !emulate {
			log.Printf("Querying serial port %s", serialport)
			serial.Connect(serialport)
		}

		serialnumber := readSerialNumber()
		protocol := readProtocol()

		for {
			data := getDataPoint()
			disp := fmt.Sprintf(`
Date: %v

Serial Number: %v
     Protocol: %v

 udc: % 8.1f V
 idc: % 8.1f A
 pdc: % 8.1f kW
 uac: % 8.1f V
 iac: % 8.1f A
 pac: % 8.1f kW
  wd: % 8.1f kWh
wtot: % 8.1f kWh
temp: % 8.1f °C
flux: % 8.1f W/m^2

Polling every %v seconds. Abort with Ctlr+C
`, data.Date, serialnumber, protocol,
				data.DC.Voltage, data.DC.Current, data.DC.Power,
				data.AC.Voltage, data.AC.Current, data.AC.Power,
				data.EnergyDay, data.EnergyTotal, data.Temperature,
				data.HeatFlux,
				pollInterval)

			area.Update(disp)
			time.Sleep(time.Second * time.Duration(pollInterval))
		}

	},
}

func getDataPoint() protocol.DataPoint {
	buff := make([]byte, 13)

	if emulate {
		buff = []byte("\x8e\x11\x82\x06\x46\x05\x06\x07\x08\x09\x0a\x0b\xa5")
	} else {
		serial.Send([]byte("\x00\x01\x02\x01\x04"))
		serial.Receive(buff)
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

func readSerialNumber() string {
	buff := make([]byte, 13)

	if emulate {
		buff = []byte("1533A5012345\x71")
	} else {
		serial.Send([]byte("\x00\x01\x08\x01\x0A"))
		serial.Receive(buff)
	}

	err := protocol.VerifyChecksum(buff)
	if err != nil {
		log.Print(err)
	}

	return string(buff[:12])
}

func readProtocol() string {
	buff := make([]byte, 13)

	if emulate {
		buff = []byte("111-23\x00\x00\x00\x00\x00\x00\x25")
	} else {
		serial.Send([]byte("\x00\x01\x09\x01\x0B"))
		serial.Receive(buff)
	}

	err := protocol.VerifyChecksum(buff)
	if err != nil {
		log.Print(err)
	}

	return string(buff[:6])
}

var cmdEmulator = &cobra.Command{
	Use:   "emulator",
	Short: "Emulate a NT5000 at the given serial port",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Emulating on port %s\n", serialport)

		serial.Connect(serialport)

		setupCloseHandler()

		buff := make([]byte, 5)

		rand.Seed(time.Now().UnixNano())

		d := protocol.DataPoint{}
		d.EnergyTotal = 1000.0 * rand.Float32()
		var lastReading int64 = 0

		for {
			err := serial.Receive(buff)
			if err != nil {
				// ignore incomplete or no data
				continue
			}
			err = protocol.VerifyChecksum(buff)
			if err != nil {
				log.Print(err)
			}

			var response []byte = nil

			switch buff[2] {
			case 0x02: // read data
				fmt.Printf("Read data\n")

				d.DC.Power = 4500.0 * rand.Float32()
				d.DC.Voltage = float32(500.0)
				d.DC.Current = d.DC.Power / d.DC.Voltage
				d.AC.Power = d.DC.Power
				d.AC.Voltage = float32(230.0)
				d.AC.Current = d.AC.Power / d.AC.Voltage
				d.Temperature = 60*rand.Float32() - 20 // between -20°C/+40°C
				d.HeatFlux = 100 * rand.Float32()

				now := time.Now().UnixMilli()
				if lastReading > 0 {
					millis := now - lastReading
					energy := d.DC.Power * float32(millis) / 1000.0 / 3600.0 / 1000.0
					d.EnergyDay += energy
					d.EnergyTotal += energy
					log.Printf("Energy %v Wh in last %v ms\n", energy, millis)
					log.Printf("Updated energyDay=%v\n", d.EnergyDay)
				}
				lastReading = now

				response = protocol.ConvertToByte(d)

			case 0x06: // read time
				fmt.Printf("Read time\n")
				response = make([]byte, 13)
				now := time.Now().Local()
				response[0] = byte(now.Year() - 2000)
				response[1] = byte(now.Month())
				response[2] = byte(now.Day())
				response[3] = byte(now.Hour())
				response[4] = byte(now.Minute())
			case 0x50: // set year
				fmt.Printf("Set year --> %v\n", int(buff[3])+2000)
			case 0x51: // set month
				fmt.Printf("Set month --> %v\n", int(buff[3]))
			case 0x52: // set day
				fmt.Printf("Set day --> %v\n", int(buff[3]))
			case 0x53: // set hour
				fmt.Printf("Set hour --> %v\n", int(buff[3]))
			case 0x54: // set minute
				fmt.Printf("Set minute --> %v\n", int(buff[3]))
			case 0x08: // read serial
				fmt.Printf("Read serial number\n")
				response = []byte("1533A5012345\x00")
			case 0x09: // read protocol + firmware
				fmt.Printf("Read protocol + firmware\n")
				response = []byte("111-23\x00\x00\x00\x00\x00\x00\x00")
			default:
				fmt.Printf("Unknown command: %v\n", buff)
			}

			if response != nil {
				if len(response) != 13 {
					log.Fatalf("Response array has length %v, expected 13\n", len(response))
				}

				protocol.CalculateChecksum(response)
				serial.Send(response)
			}

		}
	},
}

// see https://golangcode.com/handle-ctrl-c-exit-in-terminal/
func setupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Printf("Ctlr+C pressed, exiting...")
		serial.Disconnect()
		os.Exit(0)
	}()
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&serialport, "tty", "t", "/dev/ttyUSB0", "Serial port")
	rootCmd.PersistentFlags().BoolVarP(&emulate, "emulate", "e", false, "Don't use serial port at all, use fake data")

	cmdWeb.Flags().StringVarP(&port, "port", "p", "8080", "TCP port to listen on")
	cmdDatetime.Flags().BoolVarP(&settime, "set", "s", false, "Sets the date and time")
	cmdDisplay.Flags().Uint8VarP(&pollInterval, "poll", "n", 5, "Poll every n seconds")
	cmdWeb.Flags().Uint8VarP(&pollInterval, "poll", "n", 5, "Poll every n seconds")

	rootCmd.AddCommand(cmdWeb)
	rootCmd.AddCommand(cmdSerial)
	rootCmd.AddCommand(cmdDatetime)
	rootCmd.AddCommand(cmdDisplay)
	rootCmd.AddCommand(cmdEmulator)
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
