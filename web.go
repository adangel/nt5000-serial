package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"

	"github.com/pkg/browser"
)

func startWebServer(cmd *cobra.Command, args []string) {
	url := fmt.Sprintf("http://localhost:%s/", port)
	log.Printf("Starting... %v\n", url)

	if pollInterval < 1 || pollInterval > 100 {
		fmt.Printf("Invalid poll interval %v specified, using default\n", pollInterval)
		pollInterval = 5
	}

	recordData()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/display", handlerDisplay)
	http.HandleFunc("/data", handlerData)
	http.HandleFunc("/", handler)

	go func() {
		time.Sleep(time.Second * 2)
		browser.OpenURL(url)
	}()

	log.Fatal(http.ListenAndServe(":"+port, nil))
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
