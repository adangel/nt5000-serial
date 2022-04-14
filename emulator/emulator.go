package emulator

import (
	"log"
	"math/rand"
	"time"

	"github.com/adangel/nt5000-serial/protocol"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var emulatedData protocol.DataPoint = protocol.DataPoint{
	EnergyTotal: 1000.0 * rand.Float32(),
}
var lastReading int64 = 0

func ProduceDataPoint() protocol.DataPoint {
	now := time.Now()

	emulatedData.Date = now.Local()
	emulatedData.DC.Power = 4500.0 * rand.Float32()
	emulatedData.DC.Voltage = float32(500.0)
	emulatedData.DC.Current = emulatedData.DC.Power / emulatedData.DC.Voltage
	emulatedData.AC.Power = emulatedData.DC.Power
	emulatedData.AC.Voltage = float32(230.0)
	emulatedData.AC.Current = emulatedData.AC.Power / emulatedData.AC.Voltage
	emulatedData.Temperature = 60*rand.Float32() - 20 // between -20°C/+40°C
	emulatedData.HeatFlux = 100 * rand.Float32()

	if lastReading > 0 {
		millis := now.UnixMilli() - lastReading
		energy := emulatedData.DC.Power * float32(millis) / 1000.0 / 3600.0 / 1000.0
		emulatedData.EnergyDay += energy
		emulatedData.EnergyTotal += energy
		log.Printf("Energy %v Wh in last %v ms\n", energy, millis)
		log.Printf("Updated energyDay=%v\n", emulatedData.EnergyDay)
	}
	lastReading = now.UnixMilli()
	return emulatedData
}

func CurrentTimeBytes() []byte {
	data := make([]byte, 13)
	now := time.Now().Local()
	data[0] = byte(now.Year() - 2000)
	data[1] = byte(now.Month())
	data[2] = byte(now.Day())
	data[3] = byte(now.Hour())
	data[4] = byte(now.Minute())
	for i := 5; i < 12; i++ {
		data[i] = 0x0d
	}

	protocol.CalculateChecksum(data)

	return data
}
