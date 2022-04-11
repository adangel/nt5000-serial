package protocol

import "fmt"

type DataPoint struct {
	Date        string
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
	last := len(data) - 1
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
