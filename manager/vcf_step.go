
package manager

import (
  "log"
  "os"
  "io"
  "strings"
  "compress/gzip"

  "github.com/bmeg/sifter/evaluate"
  "github.com/brentp/vcfgo"
  "github.com/bmeg/sifter/schema"
)

type VCFStep struct {
  Input string `json:"input"`
  EmitAllele bool `json:"emitAllele"`
}


func (us *VCFStep) Run(task *Task) error {
	input, err := evaluate.ExpressionString(us.Input, task.Inputs)
	if err != nil {
		return err
	}
	log.Printf("Reading %s", input)
	filePath, err := task.Path(input)
	if err != nil {
		return err
	}
	fhd, err := os.Open(filePath)
	if err != nil {
		return err
	}
  defer fhd.Close()

  var hd io.Reader
	if strings.HasSuffix(input, ".gz") || strings.HasSuffix(input, ".tgz") {
		hd, err = gzip.NewReader(fhd)
		if err != nil {
			return err
		}
	} else {
    hd = fhd
  }

  rdr, err := vcfgo.NewReader(hd, false)
  if err != nil {
      return err
  }
  for {
      variant := rdr.Read()
      if variant == nil {
          break
      }
      if us.EmitAllele {
        a := schema.Allele{
          Chromosome: variant.Chromosome,
          Start: variant.Pos,
          End: variant.Pos + uint64(len(variant.Reference)),
          ReferenceBases: variant.Reference,
          AlternateBases: variant.Alternate[0],
        }
        ov, oe := a.Render()
        for _, v := range ov {
          task.EmitVertex(v)
        }
        for _, e := range oe {
          task.EmitEdge(e)
        }
      }
  }
	return nil
}
