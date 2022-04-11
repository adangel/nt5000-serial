package protocol_test

import (
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
