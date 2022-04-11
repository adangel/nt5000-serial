package protocol_test

import (
	"bytes"
	"testing"

	"github.com/adangel/nt5000-serial/protocol"
)

func TestCalculateChecksum(t *testing.T) {
	data := []byte("\x00\x01\x02\x03\x00")
	protocol.CalculateChecksum(data)
	assertChecksum(t, 0x06, data[4])

	data = []byte("\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0A\x0B\x00")
	protocol.CalculateChecksum(data)
	assertChecksum(t, 0x42, data[12])
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("\x00\x01\x02\x03\x00")
	err := protocol.VerifyChecksum(data)
	assertWrongChecksum(t, err)
	protocol.CalculateChecksum(data)
	err = protocol.VerifyChecksum(data)
	assertCorrectChecksum(t, err)

	data = []byte("\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0A\x0B\x00")
	err = protocol.VerifyChecksum(data)
	assertWrongChecksum(t, err)
	protocol.CalculateChecksum(data)
	err = protocol.VerifyChecksum(data)
	assertCorrectChecksum(t, err)
}

func TestConvert(t *testing.T) {
	data := []byte("\x8e\x11\x82\x06\x46\x05\x06\x07\x08\x09\x0a\x0b\xa5")
	point, err := protocol.Convert(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}

	err = protocol.VerifyChecksum(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(point.Date) == 0 {
		t.Fatalf("Date is missing")
	}
	assert(t, "DC.Voltage", 497.6, point.DC.Voltage)
	assert(t, "DC.Current", 1.36, point.DC.Current)
	assert(t, "DC.Power", 0.676736, point.DC.Power)
	assert(t, "AC.Voltage", 230.0, point.AC.Voltage)
	assert(t, "AC.Current", 0.71999997, point.AC.Current)
	assert(t, "AC.Power", 0.16559999, point.AC.Power)
	assert(t, "Temperature", 30, point.Temperature)
	assert(t, "HeatFlux", 30, point.HeatFlux)
	assert(t, "EnergyDay", 1.543, point.EnergyDay)
	assert(t, "EnergyTotal", 2057, point.EnergyTotal)
}

func TestConvertToByte(t *testing.T) {
	point := protocol.DataPoint{
		DC: protocol.Measurement{
			Voltage: 497.6,
			Current: 1.36,
			Power:   0.67635,
		},
		AC: protocol.Measurement{
			Voltage: 230.0,
			Current: 0.71999997,
			Power:   0.16559999,
		},
		Temperature: 30.0,
		HeatFlux:    30.0,
		EnergyDay:   1.543,
		EnergyTotal: 2057,
	}
	data := protocol.ConvertToByte(point)
	if !bytes.Equal(data, []byte("\x8e\x11\x82\x06\x46\x05\x06\x07\x08\x09\x00\x00\x90")) {
		t.Fatalf("Invalid data conversion. len=%v data=%x\n", len(data), data)
	}
}

func assert(t *testing.T, msg string, expected float32, actual float32) {
	if expected != actual {
		t.Fatalf("Wrong %s: expected=%v actual=%v\n", msg, expected, actual)
	}
}

func assertChecksum(t *testing.T, expected byte, actual byte) {
	if expected != actual {
		t.Fatalf("Wrong checksum: expected=0x%02x actual=0x%02x\n", expected, actual)
	}
}

func assertWrongChecksum(t *testing.T, err error) {
	if err == nil {
		t.Fatalf("Expected error for wrong checksum\n")
	}
}

func assertCorrectChecksum(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Expected no error for correct checksum\n")
	}
}
