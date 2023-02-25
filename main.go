package main

import (
	"fmt"
	"os"

	"github.com/bmeg/sifter/cmd"
)

func main() {
	//flag.Parse()
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
}
