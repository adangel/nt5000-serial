package protocol

import (
	"fmt"
	"time"
)

type DataPoint struct {
	Date        time.Time
	DC          Measurement
	AC          Measurement
	Temperature float32
	HeatFlux    float32
	EnergyDay   float32
	EnergyTotal float32
}

type Measurement struct {
	Voltage float32
	Current float32
	Power   float32
}

type Error struct {
	Date time.Time
	Code byte
}

func CalculateChecksum(data []byte) {
	last := len(data) - 1
	chksum := 0
	for i := 0; i < last; i++ {
		chksum += int(data[i])
	}
	chksum = chksum % 256
	data[last] = byte(chksum)
}

func VerifyChecksum(data []byte) error {
	last := 12 // the 13th byte is always the checksum
	// unless, it's the 5th
	if len(data) < last {
		last = len(data) - 1
	}
	checksum := 0
	for i := 0; i < last; i++ {
		checksum += int(data[i])
	}
	checksum = checksum % 256
	if data[last] != byte(checksum) {
		return fmt.Errorf("Invalid checksum: expected 0x%02x, got 0x%02x\n", checksum, data[last])
	}
	return nil
}

func Convert(data []byte) (DataPoint, error) {
	d := DataPoint{}
	if len(data) != 13 {
		return d, fmt.Errorf("Invalid data, expected 13 bytes, but got %d\n", len(data))
	}

	d.Date = time.Now().Local()
	d.DC.Voltage = float32(data[0])*2.8 + 100.0
	d.DC.Current = float32(data[1]) * 0.08
	d.DC.Power = (d.DC.Voltage * d.DC.Current) / 1000
	d.AC.Voltage = float32(data[2]) + 100.0
	d.AC.Current = float32(data[3]) * 0.120
	d.AC.Power = (d.AC.Voltage * d.AC.Current) / 1000
	d.Temperature = float32(data[4]) - 40.0
	d.HeatFlux = float32(data[5]) * 6.0
	d.EnergyDay = (float32(data[6])*256 + float32(data[7])) / 1000.0
	d.EnergyTotal = float32(data[8])*256 + float32(data[9])

	return d, nil
}

func ConvertToByte(d DataPoint) []byte {
	data := make([]byte, 13)

	data[0] = byte((d.DC.Voltage - 100.0) / 2.8)
	data[1] = byte(d.DC.Current / 0.08)
	data[2] = byte(d.AC.Voltage - 100.0)
	data[3] = byte(d.AC.Current / 0.120)
	data[4] = byte(d.Temperature + 40)
	data[5] = byte(d.HeatFlux / 6.0)
	data[6] = byte(d.EnergyDay * 1000.0 / 256.0)
	data[7] = byte(d.EnergyDay * 1000.0)
	data[8] = byte(d.EnergyTotal / 256.0)
	data[9] = byte(d.EnergyTotal)

	CalculateChecksum(data)
	return data
}
