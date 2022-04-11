package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/atomicgo/cursor"
	"github.com/spf13/cobra"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

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

var currentData DataPoint

func recordData() {
	go func() {
		for {

			currentData = getDataPoint()
			gaugeDCVoltage.Set(float64(currentData.DC.Voltage))
			gaugeDCCurrent.Set(float64(currentData.DC.Current))
			gaugeDCPower.Set(float64(currentData.DC.Power))
			gaugeACVoltage.Set(float64(currentData.AC.Voltage))
			gaugeACCurrent.Set(float64(currentData.AC.Current))
			gaugeACPower.Set(float64(currentData.AC.Power))
			gaugeTemperature.Set(float64(currentData.Temperature))
			gaugeHeatFlux.Set(float64(currentData.HeatFlux))
			gaugeEnergyDay.Set(float64(currentData.EnergyDay))
			gaugeEnergyTotal.Set(float64(currentData.EnergyTotal))

			time.Sleep(time.Second * time.Duration(pollInterval))
		}
	}()
}

var gaugeDCVoltage = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "nt5000_dc_voltage",
	Help: "DC Voltage in V",
})
var gaugeDCCurrent = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "nt5000_dc_current",
	Help: "DC Current in A",
})
var gaugeDCPower = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "nt5000_dc_power",
	Help: "DC Power in kW",
})
var gaugeACVoltage = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "nt5000_ac_voltage",
	Help: "AC Voltage in V",
})
var gaugeACCurrent = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "nt5000_ac_current",
	Help: "AC Current in A",
})
var gaugeACPower = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "nt5000_ac_power",
	Help: "AC Power in kW",
})
var gaugeTemperature = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "nt5000_temperature",
	Help: "Temperature in °C",
})
var gaugeHeatFlux = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "nt5000_heat_flux",
	Help: "Heat Flux in W/m^2",
})
var gaugeEnergyDay = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "nt5000_energy_day",
	Help: "Energy harvested today in kWh",
})
var gaugeEnergyTotal = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "nt5000_energy_total",
	Help: "Energy harvested total in kWh",
})

var cmdWeb = &cobra.Command{
	Use:   "web",
	Short: "start web server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("starting... http://localhost:%s/\n", port)
		fmt.Printf("args: %v\n", args)

		if pollInterval < 1 || pollInterval > 100 {
			fmt.Printf("Invalid poll interval %v specified, using default\n", pollInterval)
			pollInterval = 5
		}

		recordData()

		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/display", handlerDisplay)
		http.HandleFunc("/data", handlerData)
		http.HandleFunc("/", handler)
		log.Fatal(http.ListenAndServe(":"+port, nil))
	},
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
			buff[4] = byte((1 + 0x50 + int(buff[3])) % 256)
			serial.Send(buff)

			// month
			buff[2] = 0x51
			buff[3] = byte(now.Month())
			buff[4] = byte((1 + 0x51 + int(buff[3])) % 256)
			serial.Send(buff)

			// day
			buff[2] = 0x52
			buff[3] = byte(now.Day())
			buff[4] = byte((1 + 0x52 + int(buff[3])) % 256)
			serial.Send(buff)

			// hour
			buff[2] = 0x53
			buff[3] = byte(now.Hour())
			buff[4] = byte((1 + 0x53 + int(buff[3])) % 256)
			serial.Send(buff)

			// minute
			buff[2] = 0x54
			buff[3] = byte(now.Minute())
			buff[4] = byte((1 + 0x54 + int(buff[3])) % 256)
			serial.Send(buff)

			serial.Disconnect()
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
				serial.Connect(serialport)
				serial.Send([]byte("\x00\x01\x06\x01\x08"))
				serial.Receive(buff)
				serial.Disconnect()
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

		area := cursor.NewArea()
		area.Clear()

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

func handler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[1:]
	fmt.Fprintf(w, "<h1>%s</h1>", title)
	fmt.Fprintf(w, `
	<p><a href='/data'>data</a></p>
	<p><a href="/display">display</a></p>
	<p><a href="/metrics">metrics for prometheus</a></p>
	`)
}

type DataPoint struct {
	Date        string
	DC          measurement
	AC          measurement
	Temperature float32
	HeatFlux    float32
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
	d := currentData
	bytes, err := json.Marshal(d)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Fprintf(w, string(bytes))
}

func handlerDisplay(w http.ResponseWriter, r *http.Request) {
	d := currentData

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
<tr><td>flux</td><td>%f W/m^2</td></tr>
</table>
</body>
</html>
	`, d.Date, d.DC.Voltage, d.DC.Current, d.DC.Power, d.AC.Voltage, d.AC.Current, d.AC.Power, d.EnergyDay, d.EnergyTotal, d.Temperature,
		d.HeatFlux)
}

func getDataPoint() DataPoint {
	log.Printf("querying serial port %s", serialport)

	buff := make([]byte, 13)

	if emulate {
		buff = []byte("\x8e\x11\x82\x06\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c")
	} else {
		serial.Connect(serialport)
		serial.Send([]byte("\x00\x01\x02\x01\x04"))
		serial.Receive(buff)
		serial.Disconnect()
	}

	var udc float32 = float32(buff[0])*2.8 + 100.0
	var idc float32 = float32(buff[1]) * 0.08
	var uac float32 = float32(buff[2]) + 100.0
	var iac float32 = float32(buff[3]) * 0.120
	var temp float32 = float32(buff[4]) - 40.0
	var wd float32 = (float32(buff[6])*256 + float32(buff[7])) / 1000.0
	var wtot float32 = float32(buff[8])*256 + float32(buff[9])
	var flux float32 = float32(buff[5]) * 6.0

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
		HeatFlux:    flux,
		EnergyDay:   wd,
		EnergyTotal: wtot,
	}

	return d
}

func readSerialNumber() string {
	log.Printf("Querying serial port %s", serialport)

	buff := make([]byte, 13)

	if emulate {
		buff = []byte("1533A5012345\x71")
	} else {
		serial.Connect(serialport)
		serial.Send([]byte("\x00\x01\x08\x01\x0A"))
		serial.Receive(buff)
		serial.Disconnect()
	}

	chksum := 0
	for i := 0; i < 12; i++ {
		chksum += int(buff[i])
	}
	chksum = chksum % 256
	if buff[12] != byte(chksum) {
		fmt.Printf("invalid checksum. expected 0x%02x but was 0x%02x\n", chksum, buff[12])
	}

	return string(buff[:12])
}

func readProtocol() string {
	log.Printf("Querying serial port %s", serialport)

	buff := make([]byte, 13)

	if emulate {
		buff = []byte("111-23\x00\x00\x00\x00\x00\x00\x25")
	} else {
		serial.Connect(serialport)
		serial.Send([]byte("\x00\x01\x09\x01\x0B"))
		serial.Receive(buff)
		serial.Disconnect()
	}

	chksum := 0
	for i := 0; i < 12; i++ {
		chksum += int(buff[i])
	}
	chksum = chksum % 256
	if buff[12] != byte(chksum) {
		fmt.Printf("invalid checksum. expected 0x%02x but was 0x%02x\n", chksum, buff[12])
	}

	return string(buff[:6])
}

var cmdEmulator = &cobra.Command{
	Use:   "emulator",
	Short: "Emulate a NT5000 at the given serial port",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Emulating on port %s\n", serialport)

		serial.Connect(serialport)

		buff := make([]byte, 5)

		rand.Seed(time.Now().UnixNano())

		var energyDay float32 = 0
		var energyTotal float32 = 1000.0 * rand.Float32()
		var lastReading int64 = 0

		for {
			err := serial.Receive(buff)
			if err != nil {
				// ignore incomplete or no data
				continue
			}
			chksum := 0
			for i := 0; i < 4; i++ {
				chksum += int(buff[i])
			}
			chksum = chksum % 256
			if byte(chksum) != buff[4] {
				fmt.Printf("invalid checksum: expected 0x%02x, got 0x%02x\n", chksum, buff[4])
			}

			var response []byte = nil

			switch buff[2] {
			case 0x02: // read data
				fmt.Printf("Read data\n")
				response = make([]byte, 13)
				power := 4500.0 * rand.Float32()

				udc := float32(500.0)
				idc := power / udc
				response[0] = byte((udc - 100.0) / 2.8)
				response[1] = byte(idc / 0.08)

				uac := float32(230.0)
				iac := power / uac
				response[2] = byte(uac - 100.0)
				response[3] = byte(iac / 0.120)

				temp := 60*rand.Float32() - 20 // between -20°C/+40°C
				response[4] = byte(temp + 40)

				response[5] = byte(100 * rand.Float32())

				now := time.Now().UnixMilli()
				if lastReading > 0 {
					millis := now - lastReading
					energy := power * float32(millis) / 1000.0 / 3600.0 / 1000.0
					energyDay += energy
					energyTotal += energy
					fmt.Printf("Energy %v Wh in %v ms\n", energy, millis)
					fmt.Printf("Updated energyDay=%v\n", energyDay)
				}
				lastReading = now

				response[6] = byte(energyDay * 1000.0 / 256.0)
				response[7] = byte(energyDay * 1000.0)
				response[8] = byte(energyTotal / 256.0)
				response[9] = byte(energyTotal)

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
				fmt.Printf("Set year\n")
				fmt.Printf(" --> %v\n", int(buff[3])+2000)
			case 0x51: // set month
				fmt.Printf("Set month\n")
				fmt.Printf(" --> %v\n", int(buff[3]))
			case 0x52: // set day
				fmt.Printf("Set day\n")
				fmt.Printf(" --> %v\n", int(buff[3]))
			case 0x53: // set hour
				fmt.Printf("Set hour\n")
				fmt.Printf(" --> %v\n", int(buff[3]))
			case 0x54: // set minute
				fmt.Printf("Set minute\n")
				fmt.Printf(" --> %v\n", int(buff[3]))
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
					log.Fatalf("response array has length %v, expected 13\n", len(response))
				}

				chksum = 0
				for i := 0; i < 12; i++ {
					chksum += int(response[i])
				}
				chksum = chksum % 256
				response[12] = byte(chksum)

				serial.Send(response)
			}

		}
	},
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&serialport, "tty", "t", "/dev/ttyUSB0", "serial port")
	rootCmd.PersistentFlags().BoolVarP(&emulate, "emulate", "e", false, "Don't use serial port, use fake data")

	cmdWeb.Flags().StringVarP(&port, "port", "p", "8080", "port to listen on")
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
