package main

import (
	"github.com/adangel/nt5000-serial/protocol"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

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

func recordPrometheusData(currentData protocol.DataPoint) {
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
}
