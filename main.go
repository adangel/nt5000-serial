package main

import (
	"fmt"
	"os"

	"github.com/adangel/nt5000-serial/cmd"
)

const version = "1.0.0"

func main() {
	err := cmd.Execute(version)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
