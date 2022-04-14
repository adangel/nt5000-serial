package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/adangel/nt5000-serial/emulator"
	"github.com/adangel/nt5000-serial/protocol"
	"github.com/adangel/nt5000-serial/serial"
	"github.com/adangel/nt5000-serial/web"
	"github.com/atomicgo/cursor"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nt5000-serial",
	Short: "communicate with sunways nt5000 converter via rs232",
}

var cmdWeb = &cobra.Command{
	Use:   "web",
	Short: "start web server",
	Run: func(c *cobra.Command, args []string) {
		web.StartWebServer(Port, checkAndGetPollInterval(), SerialPort, Emulate)
	},
}

var cmdSerial = &cobra.Command{
	Use:   "serial",
	Short: "list serial ports",
	Run:   serial.ListSerialPorts,
}

var cmdDatetime = &cobra.Command{
	Use:   "datetime",
	Short: "Get or set the current time",
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Using serial port %s", SerialPort)

		settime, _ := cmd.Flags().GetBool("set")

		if settime {
			now := time.Now().Local()
			log.Printf("Setting date to %s\n", now.Format(time.ANSIC))

			serial.Connect(SerialPort)

			buff := make([]byte, 5)
			buff[0] = 0x00
			buff[1] = 0xff

			// year
			buff[2] = 0x32
			buff[3] = byte(now.Year() - 2000)
			protocol.CalculateChecksum(buff)
			serial.Send(buff)

			// month
			buff[2] = 0x33
			buff[3] = byte(now.Month())
			protocol.CalculateChecksum(buff)
			serial.Send(buff)

			// day
			buff[2] = 0x34
			buff[3] = byte(now.Day())
			protocol.CalculateChecksum(buff)
			serial.Send(buff)

			// hour
			buff[2] = 0x35
			buff[3] = byte(now.Hour() + 1)
			protocol.CalculateChecksum(buff)
			serial.Send(buff)

			// minute
			buff[2] = 0x36
			buff[3] = byte(now.Minute() + 1)
			protocol.CalculateChecksum(buff)
			serial.Send(buff)

			serial.Disconnect()
		} else {
			var err error
			buff := make([]byte, 13)

			log.Println("Reading current date...")
			if Emulate {
				buff = emulator.CurrentTimeBytes()
			} else {
				serial.Connect(SerialPort)
				serial.Send([]byte("\x00\x01\x06\x01\x08"))
				buff, err = serial.Receive()
				if err != nil {
					log.Print(err)
				}
				serial.Disconnect()
			}

			err = protocol.VerifyChecksum(buff)
			if err != nil {
				log.Print(err)
			}

			t := time.Date(int(buff[0])+2000, time.Month(buff[1]), int(buff[2]), int(buff[3]), int(buff[4]), 0, 0, time.Local)
			log.Printf("Current time: %s\n", t.Local().Format(time.ANSIC))
		}
	},
}

var cmdDisplay = &cobra.Command{
	Use:   "display",
	Short: "Display current reading on the command line",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Displaying...")
		checkAndGetPollInterval()

		serial.SetupCloseHandler()

		area := cursor.NewArea()
		area.Clear()

		if !Emulate {
			log.Printf("Querying serial port %s", SerialPort)
			serial.Connect(SerialPort)
		}

		serialnumber := serial.ReadSerialNumber(Emulate)
		protocol, firmware := serial.ReadProtocolAndFirmware(Emulate)

		for {
			data := serial.GetDataPoint(Emulate)
			disp := fmt.Sprintf(`
Date: %v

Serial Number: %v
     Protocol: %v
     Firmware: %v

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
`, data.Date, serialnumber, protocol, firmware,
				data.DC.Voltage, data.DC.Current, data.DC.Power,
				data.AC.Voltage, data.AC.Current, data.AC.Power,
				data.EnergyDay, data.EnergyTotal, data.Temperature,
				data.HeatFlux,
				PollInterval)

			area.Update(disp)
			time.Sleep(time.Second * time.Duration(PollInterval))
		}

	},
}

var cmdEmulator = &cobra.Command{
	Use:   "emulator",
	Short: "Emulate a NT5000 at the given serial port",
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Emulating on port %s\n", SerialPort)

		serial.Connect(SerialPort)

		serial.SetupCloseHandler()

		buff := make([]byte, 5)

		for {
			var err error
			buff, err = serial.Receive()
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
				log.Printf("Read data\n")
				response = protocol.ConvertToByte(emulator.ProduceDataPoint())

			case 0x06: // read time
				log.Printf("Read time\n")
				response = emulator.CurrentTimeBytes()
			case 0x32: // set year
				log.Printf("Set year --> %v\n", int(buff[3])+2000)
			case 0x33: // set month
				log.Printf("Set month --> %v\n", int(buff[3]))
			case 0x34: // set day
				log.Printf("Set day --> %v\n", int(buff[3]))
			case 0x35: // set hour
				log.Printf("Set hour --> %v\n", int(buff[3])-1)
			case 0x36: // set minute
				log.Printf("Set minute --> %v\n", int(buff[3])-1)
			case 0x08: // read serial
				log.Printf("Read serial number\n")
				response = []byte("1533A5012345\x00")
			case 0x09: // read protocol + firmware
				log.Printf("Read protocol + firmware\n")
				response = []byte("111-23\x00\x00\x00\x00\x00\x00\x00")
			default:
				log.Printf("Unknown command: %v\n", buff)
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

var Port string
var SerialPort string
var Emulate bool
var PollInterval uint8

func init() {
	ports := serial.List()
	defaultPort := "/dev/ttyUSB0"
	if len(ports) > 0 {
		defaultPort = ports[0]
	}

	rootCmd.PersistentFlags().StringVarP(&SerialPort, "tty", "t", defaultPort, "Serial port")
	rootCmd.PersistentFlags().BoolVarP(&Emulate, "emulate", "e", false, "Don't use serial port at all, use fake data")

	cmdWeb.Flags().StringVarP(&Port, "port", "p", "8080", "TCP port to listen on")
	cmdDatetime.Flags().BoolP("set", "s", false, "Sets the date and time")
	cmdDisplay.Flags().Uint8VarP(&PollInterval, "poll", "n", 5, "Poll every n seconds")
	cmdWeb.Flags().Uint8VarP(&PollInterval, "poll", "n", 5, "Poll every n seconds")

	rootCmd.AddCommand(cmdWeb)
	rootCmd.AddCommand(cmdSerial)
	rootCmd.AddCommand(cmdDatetime)
	rootCmd.AddCommand(cmdDisplay)
	rootCmd.AddCommand(cmdEmulator)
}

func Execute(version string) error {
	rootCmd.Version = version
	return rootCmd.Execute()
}

func checkAndGetPollInterval() uint8 {
	if PollInterval < 1 || PollInterval > 100 {
		log.Printf("Invalid poll interval %v specified, using default\n", PollInterval)
		PollInterval = 5
	}
	return PollInterval
}
