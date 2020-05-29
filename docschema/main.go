
package main


import (
  "fmt"
  "github.com/alecthomas/jsonschema"
  "github.com/bmeg/sifter/manager"
)



func main() {
  sch := jsonschema.Reflect(&manager.Playbook{})
  out, _ := sch.MarshalJSON()
  fmt.Printf("%s\n", out)
}
