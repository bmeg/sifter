package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bmeg/sifter/cmd"
)

func main() {
	flag.Set("v", "2")
	flag.Parse()
	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
}
