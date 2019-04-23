
package manager

import (
  "fmt"
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
      //fmt.Printf("%s\t%d\t%s\t%v\n", variant.Chromosome, variant.Pos, variant.Ref(), variant.Alt())

      a := schema.Allele{
        Chromosome: variant.Chromosome,
        Start: variant.Pos,
        End: variant.Pos + uint64(len(variant.Reference)),
        ReferenceBases: variant.Reference,
        AlternateBases: variant.Alternate[0],
      }
      fmt.Printf("%#v\n", a)
      //dp, err := variant.Info().Get("DP")
      //fmt.Printf("depth: %v\n", dp.(int))
      //sample := variant.Samples[0]
      // we can get the PL field as a list (-1 is default in case of missing value)
      //PL, err := variant.GetGenotypeField(sample, "PL", -1)
      //if err == nil {
      //  fmt.Printf("%v\n", PL)
      //  _ = sample.DP
      //}
  }
  fmt.Fprintln(os.Stderr, rdr.Error())


	return nil
}
