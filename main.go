
package main

import (
  "fmt"
  "github.com/bmeg/sifter/playbook"
  "github.com/bmeg/sifter/loader"
  "flag"
  "log"
  "os"
)

func main() {
  flag.Parse()

  pb := playbook.Playbook{}

  fmt.Printf("Starting: %s\n", flag.Args()[0])

  if err := playbook.ParseFile(flag.Args()[0], &pb); err != nil {
    log.Printf("%s", err)
  }

  //fmt.Printf("%s", pb)

  for _, step := range pb.Steps {
    if step.MatrixLoad != nil {
      log.Printf("%s\n", step.Desc)
      ml := loader.MatrixLoader{}
      input, err := os.Open(step.Input)
      if err == nil {
        elemStream := ml.Load(input, step.MatrixLoad)
        for elem := range elemStream {
          log.Printf("%s", elem)
        }
        input.Close()
      } else {
        log.Printf("Error opening: %s", step.Input)
      }
    }
  }

}
