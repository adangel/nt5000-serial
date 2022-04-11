package serial

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"go.bug.st/serial"
)

func ListSerialPorts(cmd *cobra.Command, args []string) {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}
	fmt.Printf("Found %v ports:\n", len(ports))
	for _, port := range ports {
		fmt.Printf("- %v\n", port)
	}
}
