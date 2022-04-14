package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/adangel/nt5000-serial/prometheus"
	"github.com/adangel/nt5000-serial/protocol"
	"github.com/adangel/nt5000-serial/serial"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/pkg/browser"
)

var currentData protocol.DataPoint
var basicInfo struct {
	serialnumber string
	protocol     string
	firmware     string
}

func StartWebServer(port string, pollInterval uint8, serialPort string, emulate bool) {
	url := fmt.Sprintf("http://localhost:%s/", port)
	log.Printf("Starting... %v\n", url)

	if !emulate {
		log.Printf("Querying serial port %s", serialPort)
		serial.Connect(serialPort)
	}
	basicInfo.serialnumber = serial.ReadSerialNumber(emulate)
	basicInfo.protocol, basicInfo.firmware = serial.ReadProtocolAndFirmware(emulate)
	updateDataInBackground(pollInterval, emulate)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/display", handlerDisplay)
	http.HandleFunc("/data", handlerData)
	http.HandleFunc("/", handler)

	go func() {
		time.Sleep(time.Second * 2)
		browser.OpenURL(url)
	}()
	serial.SetupCloseHandler()

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func updateDataInBackground(pollInterval uint8, emulate bool) {
	go func() {
		for {
			currentData = serial.GetDataPoint(emulate)
			prometheus.RecordPrometheusData(currentData)
			time.Sleep(time.Second * time.Duration(pollInterval))
		}
	}()
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>nt5000-serial</h1>")
	fmt.Fprintf(w, `
	<p><a href="/display">Display</a></p>
	<p><a href='/data'>JSON data</a></p>
	<p><a href="/metrics">Metrics for Prometheus</a></p>
	`)
}

func handlerData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	bytes, err := json.Marshal(currentData)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Fprintf(w, string(bytes))
}

func handlerDisplay(w http.ResponseWriter, r *http.Request) {
	d := currentData

	fmt.Fprintf(w, `
<!doctype html>
<html>
<head>
	<meta name="charset" value="utf-8">
    <meta http-equiv="refresh" content="5">
	<style type="text/css">
td:first-child, dt {
	font-weight: bold;
}
table {
	border: 1px solid black;
	border-spacing: 0px;
}
td:first-child {
	border-right: 1px solid black;
}
tr:first-child td {
	border-bottom: 1px solid black;
}
	</style>
</head>
<body>
<h1>nt5000-serial</h1>
<dl>
  <dt>Serial number:</dt><dd>%s</dd>
  <dt>Protocol:</dt><dd>%s</dd>
  <dt>Firmware:</dt><dd>%s</dd>
</dl>

<table>
<tr><td>Date</td><td>%s</td></tr>
<tr><td>udc</td><td>%f V</td></tr>
<tr><td>idc</td><td>%f A</td></tr>
<tr><td>pdc</td><td>%f kW</td></tr>
<tr><td>uac</td><td>%f V</td></tr>
<tr><td>iac</td><td>%f A</td></tr>
<tr><td>pac</td><td>%f kW</td></tr>
<tr><td>temp</td><td>%f °C</td></tr>
<tr><td>flux</td><td>%f W/m^2</td></tr>
<tr><td>wd</td><td>%f kWh</td></tr>
<tr><td>wtot</td><td>%f kWh</td></tr>
</table>
</body>
</html>
	`, basicInfo.serialnumber, basicInfo.protocol, basicInfo.firmware,
		d.Date,
		d.DC.Voltage, d.DC.Current, d.DC.Power,
		d.AC.Voltage, d.AC.Current, d.AC.Power,
		d.Temperature, d.HeatFlux,
		d.EnergyDay, d.EnergyTotal,
	)
}
