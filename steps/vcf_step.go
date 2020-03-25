
package steps

import (
  "log"
  "os"
  "io"
  "strings"
  "compress/gzip"

  "github.com/bmeg/sifter/evaluate"
  "github.com/brentp/vcfgo"
  "github.com/bmeg/sifter/schema"
  "github.com/bmeg/sifter/pipeline"
)

type VCFStep struct {
  Input string `json:"input"`
  EmitAllele bool `json:"emitAllele"`
  IDTemplate     string `json:"idTemplate"`
  Label     string `json:"label"`
  EdgeLabel   string `json:"edgeLabel"`
  InfoMap   map[string]string `json:"infoMap"`
}


func (us *VCFStep) Run(task *pipeline.Task) error {
	input, err := evaluate.ExpressionString(us.Input, task.Inputs, nil)
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
      a := schema.Allele{
        Chromosome: variant.Chromosome,
        Start: variant.Pos,
        End: variant.Pos + uint64(len(variant.Reference)),
        ReferenceBases: variant.Reference,
        AlternateBases: variant.Alternate[0],
        DBSNP_RS: variant.Id_,
      }
      if us.EmitAllele {
        ov, oe := a.Render()
        for _, v := range ov {
          task.EmitVertex(v)
        }
        for _, e := range oe {
          task.EmitEdge(e)
        }
      }
      if len(us.Label) > 0 {
        data := map[string]interface{}{}
        info := variant.Info()
        for k,m := range us.InfoMap {
          if v, e := info.Get(k); e == nil {
            data[m] = v
          }
        }
        if gid, err := evaluate.ExpressionString(us.IDTemplate, task.Inputs, map[string]interface{}{"ID":variant.Id_, "INFO":data}); err == nil {
          a := schema.AlleleAnnotation{ID: gid, Label:us.Label, EdgeLabel:us.EdgeLabel, Allele:a, Data:data}
          ov, oe := a.Render()
          for _, v := range ov {
            task.EmitVertex(v)
          }
          for _, e := range oe {
            task.EmitEdge(e)
          }
        }
      }
  }
	return nil
}
