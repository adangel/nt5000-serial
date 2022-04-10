package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/atomicgo/cursor"
	"github.com/spf13/cobra"
	"go.bug.st/serial"
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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("starting... http://localhost:%s/\n", port)
		fmt.Printf("args: %v\n", args)
		http.HandleFunc("/display", handlerDisplay)
		http.HandleFunc("/data", handlerData)
		http.HandleFunc("/", handler)
		log.Fatal(http.ListenAndServe(":"+port, nil))
	},
}

var cmdSerial = &cobra.Command{
	Use:   "serial",
	Short: "list serial ports",
	Run: func(cmd *cobra.Command, args []string) {
		ports, err := serial.GetPortsList()
		if err != nil {
			log.Fatal(err)
		}
		if len(ports) == 0 {
			log.Fatal("No serial ports found!")
		}
		for _, port := range ports {
			fmt.Printf("Found port: %v\n", port)
		}
	},
}

var settime bool
var cmdDatetime = &cobra.Command{
	Use:   "datetime",
	Short: "Get or set the current time",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("args: %s\n", args)
		log.Printf("Using serial port %s", serialport)

		if settime {
			now := time.Now().Local()
			log.Printf("Setting date to %s\n", now.Format(time.ANSIC))

			mode := &serial.Mode{
				BaudRate: 9600,
				Parity:   serial.NoParity,
				DataBits: 8,
				StopBits: serial.OneStopBit,
			}

			port, err := serial.Open("/dev/ttyUSB0", mode)
			if err != nil {
				log.Fatal(err)
			}

			buff := make([]byte, 5)
			buff[0] = 0
			buff[1] = 1

			// year
			buff[2] = 0x50
			buff[3] = byte(now.Year() - 2000)
			buff[4] = byte((1 + 0x50 + int(buff[3])) % 256)
			n, err := port.Write(buff)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Sent %v bytes\n", n)

			// month
			buff[2] = 0x51
			buff[3] = byte(now.Month())
			buff[4] = byte((1 + 0x51 + int(buff[3])) % 256)
			n, err = port.Write(buff)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Sent %v bytes\n", n)

			// day
			buff[2] = 0x52
			buff[3] = byte(now.Day())
			buff[4] = byte((1 + 0x52 + int(buff[3])) % 256)
			n, err = port.Write(buff)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Sent %v bytes\n", n)

			// hour
			buff[2] = 0x53
			buff[3] = byte(now.Hour())
			buff[4] = byte((1 + 0x53 + int(buff[3])) % 256)
			n, err = port.Write(buff)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Sent %v bytes\n", n)

			// minute
			buff[2] = 0x54
			buff[3] = byte(now.Minute())
			buff[4] = byte((1 + 0x54 + int(buff[3])) % 256)
			n, err = port.Write(buff)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Sent %v bytes\n", n)

			err = port.Close()
			if err != nil {
				log.Fatal(err)
			}

		} else {
			buff := make([]byte, 13)

			log.Println("Reading current date...")
			if emulate {
				now := time.Now().Local()
				fmt.Printf("now.Year(): %v\n", now.Year())
				buff[0] = byte(now.Year() - 2000)
				buff[1] = byte(now.Month())
				buff[2] = byte(now.Day())
				buff[3] = byte(now.Hour())
				buff[4] = byte(now.Minute())
				sum := int(0)
				for i := 0; i < 12; i++ {
					sum += int(buff[i])
				}
				buff[12] = byte(sum % 256)
			} else {
				mode := &serial.Mode{
					BaudRate: 9600,
					Parity:   serial.NoParity,
					DataBits: 8,
					StopBits: serial.OneStopBit,
				}

				port, err := serial.Open("/dev/ttyUSB0", mode)
				if err != nil {
					log.Fatal(err)
				}

				n, err := port.Write([]byte("\x00\x01\x06\x01\x08"))
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("Sent %v bytes\n", n)

				port.SetReadTimeout(time.Second * 1)

				for {
					n, err := port.Read(buff)
					if err != nil {
						log.Fatal(err)
						break
					}
					if n == 0 {
						fmt.Println("\nEOF")
						break
					}
					if n == 13 {
						break
					}
				}

				err = port.Close()
				if err != nil {
					log.Fatal(err)
				}
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

		area := cursor.NewArea()
		area.Clear()

		for {
			data := getDataPoint()
			disp := fmt.Sprintf(`
Date: %v

 udc: % 8.1f V
 idc: % 8.1f A
 pdc: % 8.1f kW
 uac: % 8.1f V
 iac: % 8.1f A
 pac: % 8.1f kW
  wd: % 8.1f kWh
wtot: % 8.1f kWh
temp: % 8.1f °C

Polling every %v seconds. Abort with Ctlr+C
`, data.Date, data.DC.Voltage, data.DC.Current, data.DC.Power,
				data.AC.Voltage, data.AC.Current, data.AC.Power,
				data.EnergyDay, data.EnergyTotal, data.Temperature,
				pollInterval)

			area.Update(disp)
			time.Sleep(time.Second * time.Duration(pollInterval))
		}

	},
}

func handler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[1:]
	fmt.Fprintf(w, "<h1>%s</h1>", title)
	fmt.Fprintf(w, `
	<p><a href='/data'>data</a></p>
	<p><a href="/display">display</a></p>
	`)
}

type DataPoint struct {
	Date        string
	DC          measurement
	AC          measurement
	Temperature float32
	EnergyDay   float32
	EnergyTotal float32
}

type measurement struct {
	Voltage float32
	Current float32
	Power   float32
}

func handlerData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	d := getDataPoint()
	bytes, err := json.Marshal(d)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Fprintf(w, string(bytes))
}

func handlerDisplay(w http.ResponseWriter, r *http.Request) {
	d := getDataPoint()

	fmt.Fprintf(w, `
<html>
<head>
    <meta http-equiv="refresh" content="5">
</head>
<body>
<table border="1">
<tr><td>Date</td><td>%s</td></tr>
<tr><td>udc</td><td>%f V</td></tr>
<tr><td>idc</td><td>%f A</td></tr>
<tr><td>pdc</td><td>%f kW</td></tr>
<tr><td>uac</td><td>%f V</td></tr>
<tr><td>iac</td><td>%f A</td></tr>
<tr><td>pac</td><td>%f kW</td></tr>
<tr><td>wd</td><td>%f kWh</td></tr>
<tr><td>wtot</td><td>%f kWh</td></tr>
<tr><td>temp</td><td>%f °C</td></tr>
</table>
</body>
</html>
	`, d.Date, d.DC.Voltage, d.DC.Current, d.DC.Power, d.AC.Voltage, d.AC.Current, d.AC.Power, d.EnergyDay, d.EnergyTotal, d.Temperature)
}

func getDataPoint() DataPoint {
	log.Printf("querying serial port %s", serialport)

	buff := make([]byte, 13)

	if emulate {
		buff = []byte("\x8e\x11\x82\x06\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c")
	} else {
		mode := &serial.Mode{
			BaudRate: 9600,
			Parity:   serial.NoParity,
			DataBits: 8,
			StopBits: serial.OneStopBit,
		}

		port, err := serial.Open("/dev/ttyUSB0", mode)
		if err != nil {
			log.Fatal(err)
		}

		n, err := port.Write([]byte("\x00\x01\x02\x01\x04"))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Sent %v bytes\n", n)

		port.SetReadTimeout(time.Second * 1)

		for {
			n, err := port.Read(buff)
			if err != nil {
				log.Fatal(err)
				break
			}
			if n == 0 {
				fmt.Println("\nEOF")
				break
			}
			if n == 13 {
				break
			}
		}

		err = port.Close()
		if err != nil {
			log.Fatal(err)
		}
	}

	var udc float32 = float32(buff[0])*2.8 + 100.0
	var idc float32 = float32(buff[1]) * 0.08
	var uac float32 = float32(buff[2]) + 100.0
	var iac float32 = float32(buff[3]) * 0.120
	var temp float32 = float32(buff[4]) - 40.0
	var wd float32 = (float32(buff[6])*256 + float32(buff[7])) / 1000.0
	var wtot float32 = float32(buff[8])*256 + float32(buff[9])

	d := DataPoint{
		Date: time.Now().Local().Format(time.ANSIC),
		DC: measurement{
			Voltage: udc,
			Current: idc,
			Power:   (udc * idc) / 1000,
		},
		AC: measurement{
			Voltage: uac,
			Current: iac,
			Power:   (uac * iac) / 1000,
		},

		Temperature: temp,
		EnergyDay:   wd,
		EnergyTotal: wtot,
	}

	return d
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&serialport, "tty", "t", "/dev/ttyUSB0", "serial port")
	rootCmd.PersistentFlags().BoolVarP(&emulate, "emulate", "e", false, "Don't use serial port, use fake data")

	cmdWeb.Flags().StringVarP(&port, "port", "p", "8080", "port to listen on")
	cmdDatetime.Flags().BoolVarP(&settime, "set", "s", false, "Sets the date and time")
	cmdDisplay.Flags().Uint8VarP(&pollInterval, "poll", "n", 5, "Poll every n seconds")

	rootCmd.AddCommand(cmdWeb)
	rootCmd.AddCommand(cmdSerial)
	rootCmd.AddCommand(cmdDatetime)
	rootCmd.AddCommand(cmdDisplay)
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
