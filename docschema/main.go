package main

import (
	"fmt"

	"github.com/alecthomas/jsonschema"
	"github.com/bmeg/sifter/playbook"
)

func main() {
	sch := jsonschema.Reflect(&playbook.Playbook{})
	out, _ := sch.MarshalJSON()
	fmt.Printf("%s\n", out)
}
