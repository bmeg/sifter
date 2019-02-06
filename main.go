
package main

import (
  "fmt"
  "github.com/bmeg/sifter/playbook"
)

func main() {
  fmt.Printf("Starting\n")

  pb := playbook.Playbook{}

  playbook.ParseFile()

}
