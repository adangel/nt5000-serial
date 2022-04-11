//go:build !darwin

package serial

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"go.bug.st/serial/enumerator"
)

func ListSerialPorts(cmd *cobra.Command, args []string) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}
	fmt.Printf("Found %v ports:\n", len(ports))
	for _, port := range ports {
		fmt.Printf("- %v\n", port.Name)
		if port.IsUSB {
			fmt.Printf("   USB ID     %s:%s\n", port.VID, port.PID)
			fmt.Printf("   USB serial %s\n", port.SerialNumber)
			if len(port.Product) != 0 {
				fmt.Printf("   Product    %s\n", port.Product)
			}
		}
	}
}
